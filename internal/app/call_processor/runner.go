package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/webitel/call_audit/model"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Runner struct {
	cfg *Config
}

func NewRunner(cfg *Config) *Runner {
	return &Runner{cfg: cfg}
}

func (r *Runner) ProcessUUID(job *model.CallJob) error {
	slog.Info("Starting processing UUID", slog.String("uuid", job.Params.CallID))

	transcriptID, fromName, toName := r.fetchTranscriptInfoWithRetries(job.Params.CallID)
	if transcriptID == "" {
		slog.Warn("No transcript ID returned", slog.String("uuid", job.Params.CallID))
		return fmt.Errorf("no transcript ID found for UUID %s", job.Params.CallID)
	}

	phrases := r.getPhrases(transcriptID, job.Params.CallID)
	dialogue := buildDialogue(phrases, fromName, toName)

	if job.Params.Scorecard != 0 {
		scorecard, err := r.fetchScorecardForm(job.Params.Scorecard)
		if err != nil {
			slog.Error("Failed to fetch scorecard form", slog.String("scorecard_id", fmt.Sprint(job.Params.Scorecard)), slog.String("error", err.Error()))
			return err
		}
		prompt := buildScorecardPrompt(dialogue, scorecard, *job.Params.SaveExplanation)
		scores, comment := r.callOpenAIForScorecard(prompt, job)
		r.sendScorecardAnswers(job, scores, comment)
		return nil
	}

	summary, category := r.summarizeTranscript(dialogue, fromName, toName, job)
	r.patchSummary(job.Params.CallID, summary, category)
	return nil
}

func buildDialogue(phrases []map[string]any, from, to string) string {
	var dialogue strings.Builder
	for _, p := range phrases {
		phrase, _ := p["phrase"].(string)
		channel, _ := p["channel"].(float64)
		name := fmt.Sprintf("Спікер %d", int(channel))
		if int(channel) == 0 && from != "" {
			name = from
		} else if int(channel) == 1 && to != "" {
			name = to
		}
		dialogue.WriteString(fmt.Sprintf("%s: %s\n", name, phrase))
	}
	return dialogue.String()
}

func buildScorecardPrompt(dialogue string, form *model.ScorecardForm, explain bool) string {
	var b strings.Builder

	b.WriteString("Оціни наступний дзвінок за анкетою.")
	if explain {
		b.WriteString(" Для кожної відповіді коротко поясни, чому саме цю оцінку надано.")
	}
	b.WriteString(" Повертай результат у форматі JSON.\n\n")

	b.WriteString("Дзвінок:\n")
	b.WriteString(dialogue)
	b.WriteString("\n\nАнкета:\n")

	for i, q := range form.Questions {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, q.Question))
		switch q.Type {
		case "question_option":
			b.WriteString("Вибери один з точних варіантів:")
			for _, opt := range q.Options {
				b.WriteString(fmt.Sprintf(" - %s (%d)\n", opt.Name, opt.Score))
			}
			b.WriteString(" — будь-яке інше значення не допускається.")
		case "question_score":
			b.WriteString(fmt.Sprintf("Оцінка від %d до %d\n", q.Min, q.Max))
		}
	}

	if explain {
		b.WriteString("\nТекст пояснення формуй для кожного питання з нового рядка, прономеруй пояснення\n")
		b.WriteString("\nФормат відповіді:\n")
		b.WriteString(`{"answers":[{"score":10},...],"comment":"текст пояснення"}`)
	} else {
		b.WriteString("\nФормат відповіді:\n")
		b.WriteString(`{"answers":[{"score":10},...],"comment":"-"}`)
	}
	return b.String()
}

func (r *Runner) callOpenAIForScorecard(prompt string, job *model.CallJob) ([]int, string) {
	client := openai.NewClient(option.WithAPIKey(*job.Params.Token))

	chat, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Ти аудитор якості дзвінків. Аналізуй за формою."),
			openai.UserMessage(prompt),
		},
		//Temperature: r.cfg.OpenAITemperature,
		//ReasoningEffort: openai.ReasoningEffortLow,
	})

	if err != nil {
		slog.Error("OpenAI request failed (scorecard)", slog.String("uuid", job.Params.CallID), slog.String("error", err.Error()))
		return nil, ""
	}

	if len(chat.Choices) == 0 {
		return nil, ""
	}

	content := chat.Choices[0].Message.Content
	// remove markdown
	cleaned := strings.TrimSpace(content)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var parsed struct {
		Answers []struct {
			Score int `json:"score"`
		} `json:"answers"`
		Comment string `json:"comment"`
	}
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		slog.Error("Failed to parse OpenAI scorecard JSON", slog.String("uuid", job.Params.CallID), slog.String("raw", cleaned))
		return nil, ""
	}

	scores := make([]int, len(parsed.Answers))
	for i, a := range parsed.Answers {
		scores[i] = a.Score
	}
	return scores, parsed.Comment
}

func (r *Runner) sendScorecardAnswers(job *model.CallJob, scores []int, comment string) {
	answers := make([]map[string]int, len(scores))
	for i, score := range scores {
		answers[i] = map[string]int{"score": score}
	}

	payload := map[string]any{
		"answers": answers,
		"call_id": job.Params.CallID,
		"comment": comment,
		"form": map[string]any{
			"id": job.Params.Scorecard,
		},
	}

	jsonBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://dev.webitel.com/api/call_center/audit/rate", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webitel-Access", r.cfg.AccessToken)

	slog.Debug("POST scorecard answers", slog.Any("payload", payload))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to send scorecard answers", slog.String("uuid", job.Params.CallID), slog.String("error", err.Error()))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		slog.Info("Scorecard answers sent", slog.String("uuid", job.Params.CallID), slog.Int("status", resp.StatusCode), slog.String("body", string(body)))
	} else {
		slog.Error("Failed to send scorecard answers", slog.String("uuid", job.Params.CallID), slog.Int("status", resp.StatusCode), slog.String("body", string(body)))
	}
}

func (r *Runner) fetchTranscriptInfoWithRetries(uuid string) (string, string, string) {
	r.createTranscript(uuid)
	for i := 1; i <= r.cfg.MaxRetries; i++ {
		id, from, to := r.fetchTranscriptInfo(uuid)
		if id != "" {
			return id, from, to
		}
		slog.Info("Retrying fetchTranscriptInfo", slog.String("uuid", uuid), slog.Int("attempt", i))
		time.Sleep(time.Duration(r.cfg.DelayBetweenRetries * float64(time.Second)))
	}
	return "", "", ""
}

func (r *Runner) createTranscript(uuid string) {
	payload := map[string]any{
		"uuid": []string{uuid},
	}

	jsonBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", r.cfg.PostTranscriptURL, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webitel-Access", r.cfg.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed POST history", slog.String("uuid", uuid), slog.String("error", err.Error()))
		return
	}
	defer resp.Body.Close()
}

func (r *Runner) fetchTranscriptInfo(uuid string) (string, string, string) {
	payload := map[string]any{
		"sort": "-created_at",
		"fields": []string{
			"id", "files", "files_job", "transcripts", "variables", "has_children",
			"agent_description", "direction", "from", "to", "destination", "hold",
			"amd_ai_logs", "amd_result", "rate_id", "allow_evaluation", "form_fields",
			"parent_id", "transfer_from", "transfer_to", "created_at", "hangup_phrase",
			"user", "duration", "bill_sec", "talk_sec", "cause",
		},
		"created_at":  map[string]any{"from": 0},
		"skip_parent": true,
		"id":          []string{uuid},
	}

	jsonBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", r.cfg.PostHistoryURL, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webitel-Access", r.cfg.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed POST history", slog.String("uuid", uuid), slog.String("error", err.Error()))
		return "", "", ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", slog.String("uuid", uuid), slog.String("error", err.Error()))
		return "", "", ""
	}

	var result struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error("Failed to parse response JSON", slog.String("uuid", uuid), slog.String("error", err.Error()))
		return "", "", ""
	}

	if len(result.Items) == 0 {
		return "", "", ""
	}

	item := result.Items[0]
	transcripts, _ := item["transcripts"].([]any)
	var transcriptID string
	if len(transcripts) > 0 {
		if tr, ok := transcripts[0].(map[string]any); ok {
			transcriptID, _ = tr["id"].(string)
		}
	}

	fromName := ""
	if from, ok := item["from"].(map[string]any); ok {
		fromName, _ = from["name"].(string)
	}

	toName := ""
	if to, ok := item["to"].(map[string]any); ok {
		toName, _ = to["name"].(string)
	}

	return transcriptID, fromName, toName
}

func (r *Runner) getPhrases(transcriptID, uuid string) []map[string]any {
	url := strings.Replace(r.cfg.GetPhrasesURLTemplate, "{id}", transcriptID, 1)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webitel-Access", r.cfg.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed GET phrases", slog.String("uuid", uuid), slog.String("error", err.Error()))
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read phrases body", slog.String("uuid", uuid), slog.String("error", err.Error()))
		return nil
	}

	var result struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error("Failed to unmarshal phrases", slog.String("uuid", uuid), slog.String("error", err.Error()))
		return nil
	}

	return result.Items
}

func (r *Runner) summarizeTranscript(dialogue, from, to string, job *model.CallJob) (string, string) {

	prompt := ""
	if job.Params.DefaultPrompt != nil && *job.Params.DefaultPrompt != "" {
		prompt = fmt.Sprintf("\nОсь розмова:\n%s\n\nВідповідь повертай у форматі:\nSummary: <текст>\nCategory: <одна з %s>",
			dialogue,
			strings.Join(r.cfg.OpenAICategories, ", "))
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	openaiClient := openai.NewClient(option.WithAPIKey(*job.Params.Token))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Create the chat completion request
	chatCompletion, err := openaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Ти класифікатор дзвінків та узагальнювач."),
			openai.UserMessage(prompt),
		},
	})
	if err != nil {
		slog.Error("OpenAI request failed (summary)", slog.String("uuid", job.Params.CallID), slog.String("error", err.Error()))
		return "", ""
	}
	// Check if we got a valid response
	if len(chatCompletion.Choices) == 0 {
		slog.Error("OpenAI response has no choices", slog.String("uuid", job.Params.CallID))
		return "", ""
	}

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	slog.Error("OpenAI JSON parse error", slog.String("uuid", job.Params.CallID), slog.String("error", err.Error()))
	// 	return "", ""
	// }

	if len(result.Choices) == 0 {
		return "", ""
	}

	content := result.Choices[0].Message.Content
	lines := strings.Split(content, "\n")
	summary := ""
	category := ""
	for _, line := range lines {
		l := strings.ToLower(line)
		if strings.HasPrefix(l, "summary:") {
			summary = strings.TrimSpace(strings.TrimPrefix(line, "Summary:"))
		}
		if strings.HasPrefix(l, "category:") {
			category = strings.TrimSpace(strings.TrimPrefix(line, "Category:"))
		}
	}
	return summary, category
}

func (r *Runner) fetchScorecardForm(scorecardID int) (*model.ScorecardForm, error) {
	url := fmt.Sprintf("https://dev.webitel.com/api/call_center/audit/forms/%d", scorecardID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Failed to create GET request for scorecard form", slog.String("scorecard_id", fmt.Sprint(scorecardID)), slog.String("error", err.Error()))
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webitel-Access", r.cfg.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("GET request for scorecard form failed", slog.String("scorecard_id", fmt.Sprint(scorecardID)), slog.String("error", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("Non-200 status code for scorecard form", slog.String("scorecard_id", fmt.Sprint(scorecardID)), slog.Int("status_code", resp.StatusCode), slog.String("body", string(body)))
		return nil, fmt.Errorf("non-200 status: %d, body: %s", resp.StatusCode, body)
	}

	var form model.ScorecardForm
	if err := json.NewDecoder(resp.Body).Decode(&form); err != nil {
		slog.Error("Failed to decode scorecard form response", slog.String("scorecard_id", fmt.Sprint(scorecardID)), slog.String("error", err.Error()))
		return nil, err
	}
	return &form, nil
}

func (r *Runner) patchSummary(uuid, summary, category string) {
	url := strings.Replace(r.cfg.PatchHistoryURLTemplate, "{uuid}", uuid, 1)

	variables := []map[string]string{
		{"name": "aiCategory", "value": category},
		{"name": "aiSummary", "value": summary},
	}

	payload := map[string]any{
		"hide_missed": true,
		"variables":   variables,
	}

	jsonBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webitel-Access", r.cfg.AccessToken)

	slog.Debug("PATCH request", slog.String("uuid", uuid), slog.Any("payload", payload))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("PATCH request failed", slog.String("uuid", uuid), slog.String("error", err.Error()))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		slog.Info("PATCH successful", slog.String("uuid", uuid), slog.Int("status", resp.StatusCode), slog.String("body", string(body)))
	} else {
		slog.Error("PATCH failed", slog.String("uuid", uuid), slog.Int("status", resp.StatusCode), slog.String("body", string(body)))
	}
}

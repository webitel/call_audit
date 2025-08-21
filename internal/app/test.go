package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	processor "github.com/webitel/call_audit/internal/app/call_processor"
	"github.com/webitel/call_audit/model"
	"github.com/webitel/storage/pool"

	"net/http"
	_ "net/http/pprof"
)

type Anc struct {
	pool pool.Pool
}

type AppJobTask struct {
	App *processor.App
	Job model.CallJob
}

type anyT struct {
	name int
}

var (
	source = rand.NewSource(time.Now().UnixNano())
	r      = rand.New(source)
)

func (t *AppJobTask) Execute() {
	fmt.Println(t.Job.Params.CallID, "Execute started")
	t.App.Process(&t.Job)
	fmt.Println(t.Job.Params.CallID, "Execute complete")

	//setJobCompleted(t.App, &t.Job)
}

func setJobCompleted(app *App, job *model.CallJob) {
	_, err := app.Store.ServiceStore().Execute(context.Background(), `
		UPDATE call_audit.jobs
		SET state = 3, updated_at = NOW()
		WHERE id = $1
	`, job.ID)
	if err != nil {
		slog.Error("Failed to update job state", slog.String("error", err.Error()))
		return
	}
}

func dropAllJobs(app *App) {
	_, err := app.Store.ServiceStore().Execute(context.Background(), `TRUNCATE TABLE call_audit.jobs;`)
	if err != nil {
		slog.Error("Failed to truncate jobs table", slog.String("error", err.Error()))
	}
	slog.Info("Jobs table truncated successfully")
}

func dropFinishedJobs(app *App) {
	_, err := app.Store.ServiceStore().Execute(context.Background(), `with del as (
		    delete from call_audit.jobs where state = 3
		    returning *
		), j as (
		    select rule_id, max(call_stored_at) max_call_stored_at
		    from del
		    group by 1
		)
		update call_center.cc_call_questionnaire_rule r
		    set last_stored_at = max_call_stored_at
		from j
		where r.id = j.rule_id
	`)
	if err != nil {
		slog.Error("Failed to drop finished jobs", slog.String("error", err.Error()))
		return
	}
}

func createJobs(app *App, rule model.CallQuestionnaireRule) any {
	fieldName := rule.Variable

	query := fmt.Sprintf(`
		INSERT INTO call_audit.jobs(rule_id, type, params)
		SELECT
			$1,
			2,
			json_build_object(
				'call_id', h.id,
				'file_id', f.id,
				'stored_at', h.stored_at,
				'position', row_number() OVER (ORDER BY h.stored_at),
				'min_call_duration', $2::int,
				'token', $3::text,
				'cognitive_key', $4::text,
				'default_prompt', $5::text,
				'save_explanation', $6::bool,
				'variable', $7::text,
				'scorecard', $12::int,
				'default_prompt', $13::text
			)
		FROM call_center.cc_calls_history h
		JOIN LATERAL (
			SELECT f.id
			FROM storage.files f
			WHERE f.domain_id = h.domain_id AND f.uuid = h.id::text
			LIMIT 1
		) f ON true
		WHERE h.domain_id = $8
		AND h.parent_id IS NULL
		AND h.stored_at > $9
		AND h.payload->'%s' IS NULL
		AND h.talk_sec > $2
		AND ($10::varchar IS NULL OR h.direction = $10::varchar)
		ORDER BY h.stored_at
		LIMIT (100 - $11)
	`, fieldName)

	args := []any{
		rule.Id,
		rule.MinCallDuration,
		rule.LanguageProfileToken,
		rule.CognitiveProfileToken,
		rule.DefaultPrompt,
		rule.SaveExplanation,
		rule.Variable,
		rule.DomainId,
		rule.Last,
		rule.CallDirection,
		rule.Active,
		rule.Scorecard,
		rule.DefaultPrompt,
	}

	_, err := app.Store.ServiceStore().Execute(context.Background(), query, args...)
	if err != nil {
		slog.Error("Failed to create jobs for rule",
			slog.Int("rule_id", rule.Id),
			slog.String("error", err.Error()))
		return err
	}

	return nil
}

func getRules(app *App, limit int) (*[]model.CallQuestionnaireRule, error) {
	rules, err := app.Store.ServiceStore().Array(context.Background(),
		`SELECT
			COALESCE(r.last_stored_at, r."from") AS last,
			r.id,
			r.domain_id,
			r.name,
			r.description,
			r.call_direction,
			r.language_profile,
			r.cognitive_profile,
			r."from",
			r."to",
			r.min_call_duration,
			r.variable,
			r.default_promt,
			r.save_explanation,
			r.enabled,
			r.created_at,
			r.created_by,
			r.updated_at,
			r.updated_by,
			r.last_stored_at,
			r.scorecard,
			lp.token AS language_token,
			cp.properties->>'key' AS cognitive_key,
			(
				SELECT COUNT(*)
				FROM call_audit.jobs j
				WHERE j.rule_id = r.id
			) AS active
		FROM call_audit.call_questionnaire_rule r
		LEFT JOIN storage.language_profiles lp ON r.language_profile = lp.id
		LEFT JOIN storage.cognitive_profile_services cp ON r.cognitive_profile = cp.id
		WHERE r.enabled
		AND (
				SELECT COUNT(*)
				FROM call_audit.jobs j
				WHERE j.rule_id = r.id
			) < 100
		ORDER BY last DESC;	
		`)
	if err != nil {
		slog.Error("Failed to get active rules", slog.String("error", err.Error()))
		return nil, err
	}
	if len(rules) == 0 {
		slog.Info("No active rules found")
		return nil, nil
	}
	// process rules

	var processedRules []model.CallQuestionnaireRule
	for _, ruleMap := range rules {
		ruleData, ok := ruleMap.(map[string]interface{})
		if !ok {
			slog.Error("ruleMap is not a map[string]interface{}")
			continue
		}
		lastTime, ok := ruleData["last"].(time.Time)
		if !ok {
			slog.Error("last field is not a time.Time")
			continue
		}
		id, ok := ruleData["id"].(int32)
		if !ok {
			slog.Error("id field is not an int32")
			continue
		}
		domainId, ok := ruleData["domain_id"].(int64)
		if !ok {
			slog.Error("domain_id field is not an int32")
			continue
		}
		active, ok := ruleData["active"].(int64)
		if !ok {
			slog.Error("active field is not an int64")
			continue
		}
		callDirection, ok := ruleData["call_direction"].(string)
		if !ok {
			slog.Error("call_direction field is not a string")
			continue
		}
		name, ok := ruleData["name"].(string)
		if !ok {
			slog.Error("name field is not a string")
			continue
		}
		description, ok := ruleData["description"].(string)
		if !ok {
			slog.Error("description field is not a string")

		}
		languageToken, ok := ruleData["language_token"].(string)
		if !ok {
			slog.Error("language_profile field is not a string")
			continue
		}
		cognitiveKey, ok := ruleData["cognitive_key"].(string)
		if !ok {
			slog.Error("cognitive_key field is not a string")
			continue
		}
		scorecard, ok := ruleData["scorecard"].(int32)
		if !ok {
			slog.Error("scorecard field is not an int")
			continue
		}
		saveExplanation, ok := ruleData["save_explanation"].(bool)
		if !ok {
			slog.Error("save_explanation field is not a bool")
		}
		variable, ok := ruleData["variable"].(string)
		if !ok {
			slog.Error("variable field is not a string")
		}
		minCallDuration, ok := ruleData["min_call_duration"].(int32)
		if !ok {
			slog.Error("min_call_duration field is not an int32")
		}
		defaultPrompt, ok := ruleData["default_promt"].(string)
		if !ok {
			slog.Error("default_promt field is not a string")
		}

		rule := model.CallQuestionnaireRule{
			Last:                  lastTime,
			Id:                    int(id),
			DomainId:              int(domainId),
			Active:                int32(active),
			CallDirection:         callDirection,
			Name:                  name,
			Description:           &description,
			LanguageProfileToken:  &languageToken,
			CognitiveProfileToken: &cognitiveKey,
			Scorecard:             scorecard,
			SaveExplanation:       &saveExplanation,
			Variable:              &variable,
			MinCallDuration:       &minCallDuration,
			DefaultPrompt:         &defaultPrompt,
		}
		processedRules = append(processedRules, rule)
	}
	return &processedRules, nil
}

// getJobs retrieves jobs from the database, updating their state to "in progress" (1).
// getJobs retrieves jobs from the database, updating their state to "in progress" (1).
func getJobs(app *App) ([]model.CallJob, error) {
	raw, err := app.Store.ServiceStore().Array(context.Background(), `
		UPDATE call_audit.jobs jj
		SET state = 1
		FROM (
			SELECT *
			FROM call_audit.jobs
			WHERE state = 0
			ORDER BY id
			FOR UPDATE
			LIMIT 10
		) j
		WHERE j.id = jj.id
		RETURNING *;
	`)
	if err != nil {
		slog.Error("Failed to update jobs table", slog.String("error", err.Error()))
		return nil, err
	}

	var jobs []model.CallJob
	for _, item := range raw {
		row, ok := item.(map[string]any)
		if !ok {
			slog.Warn("unexpected row type")
			continue
		}

		var job model.CallJob

		if v, ok := row["id"].(int64); ok {
			job.ID = v
		}
		if v, ok := row["rule_id"].(int64); ok {
			job.RuleID = v
		}
		if v, ok := row["type"].(int64); ok {
			job.Type = int(v)
		}
		if v, ok := row["state"].(int64); ok {
			job.State = int(v)
		}
		if v, ok := row["call_stored_at"].(time.Time); ok {
			job.CallStoredAt = v
		}

		// Handle JSONB column `params`
		var params model.JobParams
		switch val := row["params"].(type) {
		case map[string]any:
			jsonBytes, _ := json.Marshal(val)
			_ = json.Unmarshal(jsonBytes, &params)
		case string:
			_ = json.Unmarshal([]byte(val), &params)
		case []byte:
			_ = json.Unmarshal(val, &params)
		}
		job.Params = params

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func StartJobs(app *App) {

	slog.Info("Start jobs execution")
	dropAllJobs(app)

	//Get rules and create jobs
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			rules, err := getRules(app, 100)
			if err != nil {
				slog.Error("Failed to get rules", slog.String("error", err.Error()))
				continue
			}

			for _, rule := range *rules {
				createJobs(app, rule)
				slog.Info("Created jobs for rule",
					slog.Int64("rule_id", int64(rule.Id)),
					slog.Int64("active", int64(rule.Active)),
					slog.String("last_stored_at", rule.Last.String()))
			}
		}
	}()

	// Start a goroutine to drop finished jobs every second
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			dropFinishedJobs(app)
		}
	}()

	// Create a pool with 100 workers and a queue size of 10
	p := pool.NewPool(100, 10)

	// Register the worker function
	go func() {
		cfg := processor.LoadConfig()
		procApp := processor.NewApp(cfg)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			jobs, err := getJobs(app)
			if err != nil {
				slog.Error("Failed to get jobs", slog.String("error", err.Error()))
				continue
			}

			for _, job := range jobs {
				p.Exec(&AppJobTask{
					App: procApp,
					Job: job,
				})
				slog.Info("Submitted job", slog.String("uuid", job.Params.CallID))
			}
		}
	}()

}

func setDebug() {
	//debug.SetGCPercent(-1)

	go func() {
		slog.Info("Start debug server on http://localhost:8090/debug/pprof/")
		err := http.ListenAndServe(":8090", nil)
		if err != nil {
			slog.Error(err.Error())
		}
	}()

}

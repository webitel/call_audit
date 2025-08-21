package processor

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AccessToken             string
	OpenAIApiKey            string
	OpenAIPrompt            string
	OpenAICategories        []string
	PostTranscriptURL       string
	PostHistoryURL          string
	GetPhrasesURLTemplate   string
	PatchHistoryURLTemplate string
	DelayBeforePost         float64
	DelayBetweenRetries     float64
	MaxRetries              int
	OpenAITemperature       float64
}

func LoadConfig() *Config {
	parseFloat := func(env string, def float64) float64 {
		val, err := strconv.ParseFloat(os.Getenv(env), 64)
		if err != nil {
			return def
		}
		return val
	}
	parseInt := func(env string, def int) int {
		val, err := strconv.Atoi(os.Getenv(env))
		if err != nil {
			return def
		}
		return val
	}

	return &Config{
		AccessToken:             os.Getenv("ACCESS_TOKEN"),
		OpenAIApiKey:            os.Getenv("OPENAI_API_KEY"),
		OpenAIPrompt:            os.Getenv("OPENAI_PROMPT"),
		OpenAICategories:        split(os.Getenv("OPENAI_CATEGORIES")),
		PostTranscriptURL:       os.Getenv("POST_TRANSCRIPT_URL"),
		PostHistoryURL:          os.Getenv("POST_HISTORY_URL"),
		GetPhrasesURLTemplate:   os.Getenv("GET_URL_TEMPLATE"),
		PatchHistoryURLTemplate: os.Getenv("PATCH_HISTORY_URL_TEMPLATE"),
		DelayBeforePost:         parseFloat("DELAY_BEFORE_POST", 1.0),
		DelayBetweenRetries:     parseFloat("DELAY_BETWEEN_RETRIES", 10.0),
		MaxRetries:              parseInt("MAX_RETRIES", 5),
		OpenAITemperature:       parseFloat("OPENAI_TEMPERATURE", 0.3),
	}
}

func split(raw string) []string {
	// comma-split categories
	var out []string
	for _, s := range strings.Split(raw, ",") {
		if s := strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

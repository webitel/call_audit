package model

import (
	"time"
)

type CallQuestionnaireRule struct {
	Id                    int        `db:"id"`
	DomainId              int        `db:"domain_id"`
	CreatedAt             time.Time  `db:"created_at"`
	CreatedBy             int64      `db:"created_by"`
	UpdatedAt             time.Time  `db:"updated_at"`
	UpdatedBy             int64      `db:"updated_by"`
	Last                  time.Time  `db:"last"`
	LastStoredAt          *time.Time `db:"last_stored_at"`
	Enabled               bool       `db:"enabled"`
	Name                  string     `db:"name"`
	Description           *string    `db:"description"`
	CallDirection         string     `db:"call_direction"`
	LanguageProfile       int        `db:"language_profile"`  // ID
	CognitiveProfile      int        `db:"cognitive_profile"` // ID
	From                  time.Time  `db:"from"`
	To                    *time.Time `db:"to"`
	MinCallDuration       *int32     `db:"min_call_duration"`
	Variable              *string    `db:"variable"`
	DefaultPrompt         *string    `db:"default_promt"`
	SaveExplanation       *bool      `db:"save_explanation"`
	LanguageProfileToken  *string    `db:"language_token"` // from join
	CognitiveProfileToken *string    `db:"cognitive_key"`  // from join
	Active                int32      `db:"active"`
	Scorecard             int32      `db:"scorecard"` // ID of the scorecard form
}

type CallJob struct {
	ID           int64     `db:"id"`
	RuleID       int64     `db:"rule_id"`
	Type         int       `db:"type"`
	Params       JobParams `db:"params"` // JSONB
	State        int       `db:"state"`
	CallStoredAt time.Time `db:"call_stored_at"`
}

type JobParams struct {
	CallID          string     `json:"call_id" db:"call_id"`
	FileID          int64      `json:"file_id" db:"file_id"`
	Position        int        `json:"position" db:"position"`
	StoredAt        time.Time  `json:"stored_at" db:"stored_at"`
	From            time.Time  `json:"from" db:"from"`
	To              *time.Time `json:"to,omitempty" db:"to"`
	CallDirection   string     `json:"call_direction,omitempty" db:"call_direction"`
	MinCallDuration *int       `json:"min_call_duration,omitempty"`
	Token           *string    `json:"token,omitempty"`
	CognitiveKey    *string    `json:"cognitive_key,omitempty"`
	DefaultPrompt   *string    `json:"default_prompt,omitempty"`
	SaveExplanation *bool      `json:"save_explanation,omitempty"`
	Variable        *string    `json:"variable,omitempty"`
	Scorecard       int        `json:"scorecard,omitempty"`
}

type ScorecardForm struct {
	ID        int                 `json:"id"`
	Name      string              `json:"name"`
	Questions []ScorecardQuestion `json:"questions"`
}

type ScorecardQuestion struct {
	Type     string            `json:"type"`
	Required bool              `json:"required"`
	Question string            `json:"question"`
	Options  []ScorecardOption `json:"options,omitempty"`
	Min      int               `json:"min,omitempty"`
	Max      int               `json:"max,omitempty"`
}

type ScorecardOption struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

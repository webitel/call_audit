package model

import (
	"time"
)

type LanguageProfile struct {
	ID        int       `db:"id"`
	DomainID  int       `db:"domain_id"`
	CreatedAt time.Time `db:"created_at"`
	CreatedBy int64     `db:"created_by"`
	UpdatedAt time.Time `db:"updated_at"`
	UpdatedBy int64     `db:"updated_by"`
	Name      string    `db:"name"`
	Token     *string   `db:"token"` // nullable
	Type      int       `db:"type"`
}

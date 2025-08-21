package postgres

import (
	"context"
	"log/slog"

	conf "github.com/webitel/call_audit/config"
	dberr "github.com/webitel/call_audit/internal/errors"
	"github.com/webitel/call_audit/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is the struct implementing the Store interface.
type Store struct {
	//------------call_audit stores ------------ ----//
	languageProfilesStore      store.LanguageProfileStore
	callQuestionnaireRuleStore store.CallQuestionnaireRuleStore

	serviceStore store.ServiceStore
	config       *conf.DatabaseConfig
	conn         *pgxpool.Pool
}

func New(config *conf.DatabaseConfig) *Store {
	return &Store{config: config}
}

func (s *Store) LanguageProfiles() store.LanguageProfileStore {
	if s.languageProfilesStore == nil {
		lps := NewLanguageProfileStore(s)

		s.languageProfilesStore = lps
	}
	return s.languageProfilesStore
}

func (s *Store) CallQuestionnaireRules() store.CallQuestionnaireRuleStore {
	if s.callQuestionnaireRuleStore == nil {
		cqrs := NewCallQuestionnaireRuleStore(s)
		s.callQuestionnaireRuleStore = cqrs
	}
	return s.callQuestionnaireRuleStore
}

func (s *Store) ServiceStore() store.ServiceStore {
	if s.serviceStore == nil {
		s.serviceStore = NewServiceStore(s)
	}
	return s.serviceStore
}

// Database returns the database connection or a custom error if it is not opened.
func (s *Store) Database() (*pgxpool.Pool, *dberr.DBError) { // Return custom DB error
	if s.conn == nil {
		return nil, dberr.NewDBError("store.database.check.bad_arguments", "database connection is not opened")
	}
	return s.conn, nil
}

// Open establishes a connection to the database and returns a custom error if it fails.
func (s *Store) Open() *dberr.DBError {
	config, err := pgxpool.ParseConfig(s.config.Url)
	if err != nil {
		return dberr.NewDBError("store.open.parse_config.fail", err.Error())
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return dberr.NewDBError("store.open.connect.fail", err.Error())
	}
	s.conn = conn
	slog.Debug("call_audit.store.connection_opened", slog.String("message", "postgres: connection opened"))
	return nil
}

// Close closes the database connection and returns a custom error if it fails.
func (s *Store) Close() *dberr.DBError {
	if s.conn != nil {
		s.conn.Close()
		slog.Debug("call_audit.store.connection_closed", slog.String("message", "postgres: connection closed"))
		s.conn = nil
	}
	return nil
}

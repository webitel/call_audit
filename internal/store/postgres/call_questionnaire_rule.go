package postgres

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	cr "github.com/VoroniakPavlo/call_audit/api/call_audit"
	dberr "github.com/VoroniakPavlo/call_audit/internal/errors"
	"github.com/VoroniakPavlo/call_audit/internal/store/util"
	options "github.com/VoroniakPavlo/call_audit/model/options"
)

type QuestionnaireRuleScan func(rule *cr.CallQuestionnaireRule) any

const (
	cqrLeft                           = "cqr"
	closeQuestionnaireRuleDefaultSort = "name"
)

// CallQuestionnaireRule provides methods to interact with call questionnaire rules in the database.
type CallQuestionnaireRuleStore struct {
	storage *Store
}

// Create implements store.CallQuestionnaireRuleStore.
func (c *CallQuestionnaireRuleStore) Create(ctx context.Context, rule *cr.CallQuestionnaireRule) (*cr.CallQuestionnaireRule, error) {
	panic("unimplemented")
}

// Delete implements store.CallQuestionnaireRuleStore.
func (c *CallQuestionnaireRuleStore) Delete(ctx context.Context, id int64) error {
	panic("unimplemented")
}

// Get implements store.CallQuestionnaireRuleStore.
func (c *CallQuestionnaireRuleStore) Get(ctx context.Context, id int64) (*cr.CallQuestionnaireRule, error) {
	panic("unimplemented")
}

// List implements store.CallQuestionnaireRuleStore.
func (c *CallQuestionnaireRuleStore) List(ctx context.Context) (*cr.CallQuestionnaireRuleList, error) {
	_, dbErr := c.storage.Database()
	if dbErr != nil {
		return nil, dberr.NewDBInternalError("postgres.call_questionnaire_rule.list.database_connection_error", dbErr)
	}

	return &cr.CallQuestionnaireRuleList{}, nil
}

// Update implements store.CallQuestionnaireRuleStore.
func (c *CallQuestionnaireRuleStore) Update(ctx context.Context, rule *cr.CallQuestionnaireRule) (*cr.CallQuestionnaireRule, error) {
	panic("unimplemented")
}

// NewCallQuestionnaireRuleStore creates a new CallQuestionnaireRuleStore.
func NewCallQuestionnaireRuleStore(storage *Store) *CallQuestionnaireRuleStore {
	return &CallQuestionnaireRuleStore{storage: storage}
}

func (s *CallQuestionnaireRuleStore) buildListCallQuestionnaireRuleQuery(
	rpc options.SearchOptions,
	callQuestionnaireRuleId int64,
) (sq.SelectBuilder, []QuestionnaireRuleScan, error) {
	queryBuilder := sq.Select().
		From("call_audit.call_questionnaire_rule AS cqr").
		Where(sq.Eq{"cqr.domain": rpc.GetAuthOpts().GetDomainId()}).
		PlaceholderFormat(sq.Dollar)

	// Add ID filter if provided
	if len(rpc.GetIDs()) > 0 {
		queryBuilder = queryBuilder.Where(sq.Eq{"cqr.id": rpc.GetIDs()})
	}

	// -------- Apply sorting ----------
	queryBuilder = util.ApplyDefaultSorting(rpc, queryBuilder, closeQuestionnaireRuleDefaultSort)

	// ---------Apply paging based on Search Opts ( page ; size ) -----------------
	queryBuilder = util.ApplyPaging(rpc.GetPage(), rpc.GetSize(), queryBuilder)

	// Add select columns and scan plan for requested fields
	queryBuilder, plan, err := buildQuestionnaireRuleSelectColumnsAndPlan(queryBuilder, rpc.GetFields())
	if err != nil {
		return sq.SelectBuilder{}, nil, dberr.NewDBInternalError("postgres.questionnaire_rule.search.query_build_error", err)
	}

	return queryBuilder, plan, nil
}

func buildQuestionnaireRuleSelectColumnsAndPlan(
	base sq.SelectBuilder,
	fields []string,
) (sq.SelectBuilder, []QuestionnaireRuleScan, error) {
	var plan []QuestionnaireRuleScan
	for _, field := range fields {
		switch field {
		case "id":
			base = base.Column(util.Ident(cqrLeft, "id"))
			plan = append(plan, func(rule *cr.CallQuestionnaireRule) any {
				return &rule.Id
			})
		case "name":
			base = base.Column(util.Ident(cqrLeft, "name"))
			plan = append(plan, func(rule *cr.CallQuestionnaireRule) any {
				return &rule.Name
			})
		case "description":
			base = base.Column(util.Ident(cqrLeft, "description"))
			plan = append(plan, func(rule *cr.CallQuestionnaireRule) any {
				return &rule.Description
			})

		case "created_at":
			base = base.Column(util.Ident(cqrLeft, "created_at"))
			plan = append(plan, func(rule *cr.CallQuestionnaireRule) any {
				return &rule.CreatedAt
			})
		case "updated_at":
			base = base.Column(util.Ident(cqrLeft, "updated_at"))
			plan = append(plan, func(rule *cr.CallQuestionnaireRule) any {
				return &rule.UpdatedAt
			})
		case "enabled":
			base = base.Column(util.Ident(cqrLeft, "enabled"))
			plan = append(plan, func(rule *cr.CallQuestionnaireRule) any {
				return &rule.Enabled
			})
		case "domain_id":
			base = base.Column(util.Ident(cqrLeft, "domain_id"))
			plan = append(plan, func(rule *cr.CallQuestionnaireRule) any {
				return &rule.DomainId
			})
		case "last_stored_at":
			base = base.Column(util.Ident(cqrLeft, "last_stored_at"))
			plan = append(plan, func(rule *cr.CallQuestionnaireRule) any {
				return &rule.LastStoredAt
			})
		default:
			return base, nil, dberr.NewDBInternalError("postgres.close_reason.unknown_field", fmt.Errorf("unknown field: %s", field))
		}
	}
	return base, plan, nil
}

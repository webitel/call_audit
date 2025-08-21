package app

import (
	"context"
	"fmt"

	pb "github.com/webitel/call_audit/api/call_audit"
)

type CallQuestionnaireRuleService struct {
	app *App
	pb.UnimplementedCallQuestionnaireRuleServiceServer
}

func NewCallQuestionnaireRuleService(app *App) (*CallQuestionnaireRuleService, error) {
	service := &CallQuestionnaireRuleService{
		app: app,
	}
	return service, nil
}

func (s *CallQuestionnaireRuleService) List(ctx context.Context, req *pb.Empty) (*pb.CallQuestionnaireRuleList, error) {
	lp, err := s.app.Store.CallQuestionnaireRules().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list call questionnaire rules: %w", err)
	}
	return lp, nil
}

func (s *CallQuestionnaireRuleService) Create(ctx context.Context, req *pb.UpsertCallQuestionnaireRuleRequest) (*pb.CallQuestionnaireRule, error) {
	ruleReq := req.GetRule()
	if ruleReq == nil {
		return nil, fmt.Errorf("request does not contain CallQuestionnaireRule")
	}
	rule, err := s.app.Store.CallQuestionnaireRules().Create(ctx, ruleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create call questionnaire rule: %w", err)
	}
	return rule, nil
}

func (s *CallQuestionnaireRuleService) Update(ctx context.Context, req *pb.UpsertCallQuestionnaireRuleRequest) (*pb.CallQuestionnaireRule, error) {
	ruleReq := req.GetRule()
	if ruleReq == nil {
		return nil, fmt.Errorf("request does not contain CallQuestionnaireRule")
	}
	rule, err := s.app.Store.CallQuestionnaireRules().Update(ctx, ruleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update call questionnaire rule: %w", err)
	}
	return rule, nil
}

func (s *CallQuestionnaireRuleService) Delete(ctx context.Context, req *pb.DeleteCallQuestionnaireRuleRequest) (*pb.CallQuestionnaireRule, error) {
	err := s.app.Store.CallQuestionnaireRules().Delete(ctx, int64(req.GetId()))
	if err != nil {
		return nil, fmt.Errorf("failed to delete call questionnaire rule: %w", err)
	}
	return &pb.CallQuestionnaireRule{Id: req.Id}, nil
}

func (s *CallQuestionnaireRuleService) Get(ctx context.Context, req *pb.GetCallQuestionnaireRuleRequest) (*pb.CallQuestionnaireRule, error) {
	rule, err := s.app.Store.CallQuestionnaireRules().Get(ctx, int64(req.Id))
	if err != nil {
		return nil, fmt.Errorf("failed to get call questionnaire rule: %w", err)
	}
	return rule, nil
}

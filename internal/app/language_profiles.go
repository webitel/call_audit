package app

import (
	"context"
	"fmt"

	pb "github.com/VoroniakPavlo/call_audit/api/call_audit"
)

type LanguageProfilesService struct {
	app *App // Define App type below or import from the correct package
	pb.UnimplementedLanguageProfileServiceServer
}

func (s *LanguageProfilesService) CreateLanguageProfile(ctx context.Context, req *pb.CreateLanguageProfileRequest) (*pb.LanguageProfile, error) {
	lp, err := s.app.Store.LanguageProfiles().Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create language profile: %w", err)
	}
	return lp, nil
}

func (s *LanguageProfilesService) GetLanguageProfile(ctx context.Context, req *pb.GetLanguageProfileRequest) (*pb.LanguageProfile, error) {
	lp, err := s.app.Store.LanguageProfiles().Get(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get language profile: %w", err)
	}
	return lp, nil
}

func (s *LanguageProfilesService) UpdateLanguageProfile(ctx context.Context, req *pb.UpdateLanguageProfileRequest) (*pb.LanguageProfile, error) {
	lp, err := s.app.Store.LanguageProfiles().Update(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update language profile: %w", err)
	}
	return lp, nil
}

func (s *LanguageProfilesService) DeleteLanguageProfile(ctx context.Context, req *pb.DeleteLanguageProfileRequest) (*pb.LanguageProfile, error) {
	err := s.app.Store.LanguageProfiles().Delete(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete language profile: %w", err)
	}
	return &pb.LanguageProfile{Id: req.Id}, nil
}

func (s *LanguageProfilesService) ListLanguageProfiles(ctx context.Context, req *pb.ListLanguageProfilesRequest) (*pb.ListLanguageProfilesResponse, error) {
	list, err := s.app.Store.LanguageProfiles().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list language profiles: %w", err)
	}
	return list, nil
}

func NewLanguageProfileService(app *App) (*LanguageProfilesService, error) {

	service := &LanguageProfilesService{
		app: app,
	}

	return service, nil
}

package postgres

import (
	"context"
	"fmt"

	pb "github.com/VoroniakPavlo/call_audit/api/call_audit"
	"github.com/webitel/storage/model"
)

// LanguageProfileStore provides methods to interact with language profiles in the database.
type LanguageProfileStore struct {
	storage *Store
}

// Create implements store.LanguageProfileStore.
func (l *LanguageProfileStore) Create(ctx context.Context, profile *pb.CreateLanguageProfileRequest) (*pb.LanguageProfile, error) {
	// Implementation goes here
	return nil, nil
}

// Delete implements store.LanguageProfileStore.
func (l *LanguageProfileStore) Delete(ctx context.Context, id int32) error {
	db, dbErr := l.storage.Database()
	if dbErr != nil {
		return model.NewCustomCodeError("store.language_profile.delete.app_error", fmt.Sprintf("Id=%v, %s", id, dbErr.Error()), 400)
	}
	if _, err := db.Exec(ctx, `delete from storage.language_profiles c where c.id=:Id`,
		map[string]interface{}{"Id": id}); err != nil {
		return model.NewCustomCodeError("store.language_profile.delete.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), 400)
	}
	return nil
}

// Get implements store.LanguageProfileStore.
func (l *LanguageProfileStore) Get(ctx context.Context, id int32) (*pb.LanguageProfile, error) {
	panic("unimplemented")
}

// List implements store.LanguageProfileStore.
func (l *LanguageProfileStore) List(ctx context.Context) (*pb.ListLanguageProfilesResponse, error) {
	panic("unimplemented")
}

// Update implements store.LanguageProfileStore.
func (l *LanguageProfileStore) Update(ctx context.Context, profile *pb.UpdateLanguageProfileRequest) (*pb.LanguageProfile, error) {
	a, _ := l.storage.Database()
	a.Exec(ctx, "SELECT 1") // Example query to check connection
	panic("unimplemented")
}

// NewLanguageProfileStore creates a new LanguageProfileStore.
func NewLanguageProfileStore(storage *Store) *LanguageProfileStore {
	return &LanguageProfileStore{storage: storage}
}

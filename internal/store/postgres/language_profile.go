package postgres

import (
	"context"
	"fmt"

	pb "github.com/webitel/call_audit/api/call_audit"
	"github.com/webitel/storage/model"
)

// LanguageProfileStore provides methods to interact with language profiles in the database.
type LanguageProfileStore struct {
	storage *Store
}

// Create implements store.LanguageProfileStore.
func (l *LanguageProfileStore) Create(ctx context.Context, req *pb.CreateLanguageProfileRequest) (*pb.LanguageProfile, error) {
	db, dbErr := l.storage.Database()
	if dbErr != nil {
		return nil, model.NewCustomCodeError("store.language_profile.create.app_error", dbErr.Error(), 500)
	}

	// TODO: derive these from ctx/request
	domainID := req.GetDomainId() // or from auth context
	userID := req.GetCreatedBy()  // who creates/updates
	name := req.GetName()
	token := req.GetToken()     // optional
	typ := int32(req.GetType()) // required

	// Insert with positional params and explicit RETURNING list
	const q = `
        INSERT INTO storage.language_profiles
            (domain_id, created_by, updated_by, name, token, "type")
        VALUES
            ($1,        $2,         $2,        $3,   $4,    $5)
        RETURNING
            id, domain_id, created_at, created_by, updated_at, updated_by, name, token, "type"
    `

	var lp pb.LanguageProfile
	// Adjust Scan order to your proto struct fields or scan into temps then map
	err := db.QueryRow(ctx, q,
		domainID, userID, name, token, typ,
	).Scan(
		&lp.Id,
		&lp.DomainId,
		&lp.CreatedAt, 
		&lp.CreatedBy,
		&lp.UpdatedAt,
		&lp.UpdatedBy,
		&lp.Name,
		&lp.Token,
		&lp.Type,
	)
	if err != nil {
		return nil, model.NewCustomCodeError("store.language_profile.create.app_error", err.Error(), 500)
	}
	return &lp, nil
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
	db, dbErr := l.storage.Database()
	if dbErr != nil {
		return nil, model.NewCustomCodeError("store.language_profile.get.app_error", fmt.Sprintf("Id=%v, %s", id, dbErr.Error()), 400)
	}
	var profile pb.LanguageProfile
	if err := db.QueryRow(ctx, `select * from storage.language_profiles c where c.id=:Id`,
		map[string]interface{}{"Id": id}).Scan(&profile); err != nil {
		return nil, model.NewCustomCodeError("store.language_profile.get.app_error", fmt.Sprintf("Id=%v, %s", id, err.Error()), 400)
	}
	return &profile, nil
}

// List implements store.LanguageProfileStore.
func (l *LanguageProfileStore) List(ctx context.Context) (*pb.ListLanguageProfilesResponse, error) {
	db, dbErr := l.storage.Database()
	if dbErr != nil {
		return nil, model.NewCustomCodeError("store.language_profile.list.app_error", dbErr.Error(), 500)
	}

	var profiles []*pb.LanguageProfile

	rows, err := db.Query(ctx, `select * from storage.language_profiles`)
	if err != nil {
		return nil, model.NewCustomCodeError("store.language_profile.list.app_error", err.Error(), 500)
	}
	defer rows.Close()

	for rows.Next() {
		var profile pb.LanguageProfile
		if err := rows.Scan(&profile); err != nil {
			return nil, model.NewCustomCodeError("store.language_profile.list.app_error", err.Error(), 500)
		}
		profiles = append(profiles, &profile)
	}

	return &pb.ListLanguageProfilesResponse{Profiles: profiles}, nil
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

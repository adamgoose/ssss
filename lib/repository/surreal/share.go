package surreal

import (
	"github.com/adamgoose/ssss/lib/model"
	"github.com/defval/di"
	"github.com/surrealdb/surrealdb.go"
)

type SurrealShareRepository struct {
	di.Inject
	DB *surrealdb.DB
}

func (r SurrealShareRepository) MineForSecret(secretID string, userID string) ([]model.Share, error) {
	data, err := r.DB.Query("SELECT * FROM shares WHERE secret = $id AND user = $user", map[string]interface{}{
		"id":   secretID,
		"user": userID,
	})
	if err != nil {
		return nil, err
	}

	result := []surrealdb.RawQuery[[]model.Share]{}
	if err := surrealdb.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result[0].Result, nil
}

func (r SurrealShareRepository) Create(share *model.Share) (*model.Share, error) {
	data, err := r.DB.Create("shares", share)
	if err != nil {
		return nil, err
	}

	ns := make([]model.Share, 1)
	if err := surrealdb.Unmarshal(data, &ns); err != nil {
		return nil, err
	}

	data, err = r.DB.Query("RELATE $secret->split_into->$share", map[string]interface{}{
		"share":  ns[0].ID,
		"secret": share.Secret,
	})
	if err != nil {
		return nil, err
	}

	data, err = r.DB.Query("RELATE $user->signed->$share", map[string]interface{}{
		"user":  share.User,
		"share": ns[0].ID,
	})
	if err != nil {
		return nil, err
	}

	return &ns[0], nil
}

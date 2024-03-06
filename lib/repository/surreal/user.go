package surreal

import (
	"github.com/adamgoose/ssss/lib/model"
	"github.com/defval/di"
	"github.com/surrealdb/surrealdb.go"
)

type SurrealUserRepository struct {
	di.Inject
	DB *surrealdb.DB
}

func (r SurrealUserRepository) Upsert(user *model.User) (*model.User, error) {
	data, err := r.DB.Query(`
		INSERT INTO users (id, username, public_key, first_seen, last_seen)
		VALUES ([$username, $public_key], $username, $public_key, time::now(), time::now())
		ON DUPLICATE KEY UPDATE last_seen = time::now()
  `, map[string]interface{}{
		"username":   user.Username,
		"public_key": user.PublicKey,
	})
	if err != nil {
		return nil, err
	}

	result := []surrealdb.RawQuery[[]model.User]{}
	if err := surrealdb.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result[0].Result[0], nil
}

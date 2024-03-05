package repository

import (
	"github.com/adamgoose/ssss/lib/model"
	"github.com/defval/di"
	"github.com/surrealdb/surrealdb.go"
)

type SecretRepository interface {
	Get(id string) (*model.Secret, error)
	Create(secret *model.Secret) (*model.Secret, error)
	Update(secret *model.Secret) error
}

var _ SecretRepository = SurrealSecretRepository{}

type SurrealSecretRepository struct {
	di.Inject
	DB *surrealdb.DB
}

// Get implements SecretRepository.
func (r SurrealSecretRepository) Get(id string) (*model.Secret, error) {
	data, err := r.DB.Select("secrets:" + id)
	if err != nil {
		return nil, err
	}

	secret := model.Secret{}
	if err := surrealdb.Unmarshal(data, &secret); err != nil {
		return nil, err
	}

	return &secret, nil
}

// Create implements SecretRepository.
func (r SurrealSecretRepository) Create(secret *model.Secret) (*model.Secret, error) {
	data, err := r.DB.Create("secrets", secret)
	if err != nil {
		return nil, err
	}

	ns := make([]model.Secret, 1)
	if err := surrealdb.Unmarshal(data, &ns); err != nil {
		return nil, err
	}

	return &ns[0], nil
}

// Update implements SecretRepository.
func (r SurrealSecretRepository) Update(secret *model.Secret) error {
	_, err := r.DB.Update(secret.ID, secret)
	return err
}

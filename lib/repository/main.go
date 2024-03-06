package repository

import "github.com/adamgoose/ssss/lib/model"

type Repository interface {
	User() UserRepository
	Share() ShareRepository
	Secret() SecretRepository
}
type UserRepository interface {
	Upsert(user *model.User) (*model.User, error)
}

type ShareRepository interface {
	MineForSecret(secretID string, userID string) ([]model.Share, error)
	Create(share *model.Share) (*model.Share, error)
}

type SecretRepository interface {
	Get(id string) (*model.Secret, error)
	Mine(userID string) ([]model.Secret, error)
	Create(secret *model.Secret) (*model.Secret, error)
	Update(secret *model.Secret) error
}

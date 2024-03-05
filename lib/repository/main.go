package repository

import (
	"github.com/adamgoose/ssss/lib"
)

type Repository interface {
	User() UserRepository
	Share() ShareRepository
	Secret() SecretRepository
}

var _ Repository = SurrealRepository{}

type SurrealRepository struct {
	//
}

func (r SurrealRepository) User() UserRepository {
	return lib.MustAutoResolve[SurrealUserRepository]()
}

func (r SurrealRepository) Share() ShareRepository {
	return lib.MustAutoResolve[SurrealShareRepository]()
}

func (r SurrealRepository) Secret() SecretRepository {
	return lib.MustAutoResolve[SurrealSecretRepository]()
}

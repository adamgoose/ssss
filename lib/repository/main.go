package repository

import (
	"github.com/adamgoose/ssss/lib"
)

type Repository interface {
	Secret() SecretRepository
	Share() ShareRepository
}

var _ Repository = SurrealRepository{}

type SurrealRepository struct {
	//
}

func (r SurrealRepository) Secret() SecretRepository {
	return lib.MustAutoResolve[SurrealSecretRepository]()
}

func (r SurrealRepository) Share() ShareRepository {
	return lib.MustAutoResolve[SurrealShareRepository]()
}

package surreal

import (
	"github.com/adamgoose/ssss/lib"
	"github.com/adamgoose/ssss/lib/repository"
)

var _ repository.Repository = SurrealRepository{}

type SurrealRepository struct {
	//
}

func (r SurrealRepository) User() repository.UserRepository {
	return lib.MustAutoResolve[SurrealUserRepository]()
}

func (r SurrealRepository) Share() repository.ShareRepository {
	return lib.MustAutoResolve[SurrealShareRepository]()
}

func (r SurrealRepository) Secret() repository.SecretRepository {
	return lib.MustAutoResolve[SurrealSecretRepository]()
}

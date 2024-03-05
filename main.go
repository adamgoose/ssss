package main

import (
	"log"

	"github.com/adamgoose/ssss/cmd"
	"github.com/adamgoose/ssss/lib"
	"github.com/adamgoose/ssss/lib/repository"
	_ "github.com/corvus-ch/shamir"
	"github.com/defval/di"
	"github.com/surrealdb/surrealdb.go"
)

func main() {
	if err := lib.Apply(
		di.Provide(func() (*surrealdb.DB, error) {
			db, err := surrealdb.New("ws://localhost:4222/rpc")
			if err != nil {
				return nil, err
			}

			if _, err = db.Use("test", "test"); err != nil {
				return nil, err
			}

			return db, nil
		}),
		di.ProvideValue(repository.SurrealRepository{}, di.As(new(repository.Repository))),
	); err != nil {
		log.Fatal(err)
	}

	if err := lib.Invoke(cmd.RunE); err != nil {
		log.Fatal(err)
	}
}

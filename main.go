package main

import (
	"log"

	"github.com/adamgoose/ssss/cmd"
	"github.com/adamgoose/ssss/lib"
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
	); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

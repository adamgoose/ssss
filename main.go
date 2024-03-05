package main

import (
	"log"

	"github.com/adamgoose/ssss/cmd"
	"github.com/adamgoose/ssss/lib"
	"github.com/adamgoose/ssss/lib/repository"
	_ "github.com/corvus-ch/shamir"
	"github.com/defval/di"
	"github.com/spf13/viper"
	"github.com/surrealdb/surrealdb.go"
)

func main() {
	if err := lib.Apply(
		di.Provide(func() (*surrealdb.DB, error) {
			db, err := surrealdb.New(viper.GetString("surrealdb_address"))
			if err != nil {
				return nil, err
			}

			if _, err = db.Signin(map[string]interface{}{
				"user": viper.GetString("surrealdb_user"),
				"pass": viper.GetString("surrealdb_pass"),
			}); err != nil {
				return nil, err
			}

			if _, err = db.Use(viper.GetString("surrealdb_ns"), viper.GetString("surrealdb_db")); err != nil {
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

func init() {
	viper.SetEnvPrefix("ssss")
	viper.AutomaticEnv()

	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", "23234")
	viper.SetDefault("host_key_path", ".ssh/id_ed25519")
	viper.SetDefault("surrealdb_address", "ws://127.0.0.1:4222/rpc")
	viper.SetDefault("surrealdb_user", "root")
	viper.SetDefault("surrealdb_pass", "root")
	viper.SetDefault("surrealdb_ns", "ssss")
	viper.SetDefault("surrealdb_db", "ssss")
}

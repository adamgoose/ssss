package cmd

import (
	"github.com/adamgoose/ssss/lib"
	"github.com/corvus-ch/shamir"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/surrealdb/surrealdb.go"
)

var combineCmd = &cobra.Command{
	Use:   "combine {id}",
	Short: "Combines shares to recover a secret",
	RunE: lib.RunE(func(args []string, db *surrealdb.DB) error {
		data, err := db.Select(args[0])
		if err != nil {
			return err
		}

		secret := new(Secret)
		if err := surrealdb.Unmarshal(data, secret); err != nil {
			return err
		}

		shares := map[byte][]byte{}
		for i, share := range secret.Shares {
			if i >= secret.Threshold {
				break
			}

			shares[share.Key] = share.Share
		}

		spew.Dump(shamir.Combine(shares))

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(combineCmd)
}

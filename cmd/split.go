package cmd

import (
	"fmt"

	"github.com/adamgoose/ssss/lib"
	"github.com/corvus-ch/shamir"
	"github.com/spf13/cobra"
	"github.com/surrealdb/surrealdb.go"
)

var (
	parts     int
	threshold int
)

type Secret struct {
	ID        string        `json:"id,omitempty"`
	Parts     int           `json:"parts"`
	Threshold int           `json:"threshold"`
	Shares    []SecretShare `json:"shares"`
}

type SecretShare struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label"`
	Key   byte   `json:"key"`
	Share []byte `json:"share"`
}

var splitCmd = &cobra.Command{
	Use:   "split {secret}",
	Short: "Splits a secret into encrypted shares",
	RunE: lib.RunE(func(args []string, db *surrealdb.DB) error {
		shares := []SecretShare{}

		rawShares, err := shamir.Split([]byte(args[0]), parts, threshold)
		if err != nil {
			return err
		}

		for key, rawShare := range rawShares {
			shares = append(shares, SecretShare{
				Label: fmt.Sprintf("%x", key),
				Key:   key,
				Share: rawShare,
			})
		}

		secret := Secret{
			Parts:     parts,
			Threshold: threshold,
			Shares:    shares,
		}

		data, err := db.Create("secret", secret)
		if err != nil {
			panic(err)
		}

		storedSecret := make([]Secret, 1)
		// storedSecret[0].Shares = make([]SecretShare, parts)
		if err := surrealdb.Unmarshal(data, &storedSecret); err != nil {
			return err
		}

		fmt.Println(storedSecret[0].ID)

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(splitCmd)

	splitCmd.Flags().IntVarP(&parts, "parts", "p", 3, "Number of parts to split the secret into")
	splitCmd.Flags().IntVarP(&threshold, "threshold", "t", 2, "Number of parts required to reconstruct the secret")
}

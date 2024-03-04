package cmd

import (
	"errors"

	"github.com/adamgoose/ssss/lib"
	"github.com/charmbracelet/ssh"
	"github.com/spf13/cobra"
	"github.com/surrealdb/surrealdb.go"
)

func NewSSHCmd(sess ssh.Session) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sssc",
		Short: "Interact with ssss over SSH.",
	}

	splitCmd := &cobra.Command{
		Use:   "split",
		Short: "Splits a secret into shares.",
		RunE: lib.RunE(func(cmd *cobra.Command, db *surrealdb.DB) error {
			// p, _ := cmd.Flags().GetInt("parts")
			// t, _ := cmd.Flags().GetInt("threshold")
			// cmd.Printf("parts: %d\nthreshold: %d\n", p, t)

			p := NewSplitProgram(sess, db)
			_, err := p.Run()

			return err
		}),
	}

	signCmd := &cobra.Command{
		Use:   "sign {id}",
		Short: "Signs a share with a passphrase.",
		Args:  cobra.ExactArgs(1),
		RunE: lib.RunE(func(cmd *cobra.Command, args []string, db *surrealdb.DB) error {
			// Lookup the secret by ID
			data, err := db.Select("secrets:" + args[0])
			if err != nil {
				return err
			}

			secret := Secret{}
			if err := surrealdb.Unmarshal(data, &secret); err != nil {
				return err
			}

			// Verify it's in a signing state
			if secret.Status != "signing" {
				return errors.New("Secret is not in a signing state.")
			}

			splitState, ok := SplitStates[secret.ID]
			if !ok {
				return errors.New("Secret is not in a signing state.")
			}

			p := NewSignProgram(sess, db, splitState)
			_, err = p.Run()

			return err
		}),
	}

	combineCmd := &cobra.Command{
		Use:   "combine {id}",
		Short: "Combines shares to recover a secret.",
		Args:  cobra.ExactArgs(1),
		RunE: lib.RunE(func(cmd *cobra.Command, args []string, db *surrealdb.DB) error {
			// Lookup the secret by ID
			data, err := db.Select("secrets:" + args[0])
			if err != nil {
				return err
			}

			secret := Secret{}
			if err := surrealdb.Unmarshal(data, &secret); err != nil {
				return err
			}

			// Verify it's in a signing state
			if secret.Status != "ready" {
				return errors.New("Secret is not in a ready state.")
			}

			cs := NewCombineState(secret.ID, secret.Threshold)
			p := NewCombineProgram(sess, db, cs)
			_, err = p.Run()
			return err
		}),
	}

	unsignCmd := &cobra.Command{
		Use:   "unsign {id}",
		Short: "Unsigns a share with a passphrase.",
		Args:  cobra.ExactArgs(1),
		RunE: lib.RunE(func(cmd *cobra.Command, args []string, db *surrealdb.DB) error {
			// Lookup the secret by ID
			data, err := db.Select("secrets:" + args[0])
			if err != nil {
				return err
			}

			secret := Secret{}
			if err := surrealdb.Unmarshal(data, &secret); err != nil {
				return err
			}

			// Verify it's in a signing state
			if secret.Status != "ready" {
				return errors.New("Secret is not in a ready state.")
			}

			cs, ok := CombineStates[secret.ID]
			if !ok {
				return errors.New("Secret is not being combined.")
			}

			// Load the shares
			data, err = db.Query("SELECT * FROM shares WHERE secret = $id AND user = $user", map[string]interface{}{
				"id":   secret.ID,
				"user": sess.Context().Value(User{}).(User).ID,
			})
			if err != nil {
				return err
			}

			result := []surrealdb.RawQuery[[]Share]{}
			if err := surrealdb.Unmarshal(data, &result); err != nil {
				return err
			}
			shares := result[0].Result

			p := NewUnsignProgram(sess, db, cs, shares)
			_, err = p.Run()
			return err
		}),
	}

	splitCmd.Flags().IntP("parts", "p", 5, "How many shares to split the secret into.")
	splitCmd.Flags().IntP("threshold", "t", 3, "How many shares are required to reconstruct the secret.")

	rootCmd.AddCommand(splitCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(combineCmd)
	rootCmd.AddCommand(unsignCmd)

	return rootCmd
}

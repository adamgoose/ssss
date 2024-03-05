package cmd

import (
	"errors"

	"github.com/adamgoose/ssss/lib"
	"github.com/adamgoose/ssss/lib/repository"
	"github.com/charmbracelet/ssh"
	"github.com/defval/di"
	"github.com/spf13/cobra"
)

func NewSSHCmd(sess ssh.Session) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sssc",
		Short: "Interact with ssss over SSH.",
	}

	splitCmd := &cobra.Command{
		Use:   "split",
		Short: "Splits a secret into shares.",
		RunE: lib.RunE(func(cmd *cobra.Command, repo repository.Repository) error {
			ioc, _ := lib.Wrap(
				di.ProvideValue(cmd),
				di.ProvideValue(sess, di.As(new(ssh.Session))),
			)

			return ioc.Invoke(RunSplitProgram)
		}),
	}

	signCmd := &cobra.Command{
		Use:   "sign {id}",
		Short: "Signs a share with a passphrase.",
		Args:  cobra.ExactArgs(1),
		RunE: lib.RunE(func(cmd *cobra.Command, args []string, repo repository.Repository) error {
			// Lookup the secret by ID
			secret, err := repo.Secret().Get(args[0])
			if err != nil {
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

			ioc, _ := lib.Wrap(
				di.ProvideValue(splitState),
				di.ProvideValue(sess, di.As(new(ssh.Session))),
			)

			return ioc.Invoke(RunSignProgram)
		}),
	}

	combineCmd := &cobra.Command{
		Use:   "combine {id}",
		Short: "Combines shares to recover a secret.",
		Args:  cobra.ExactArgs(1),
		RunE: lib.RunE(func(cmd *cobra.Command, args []string, repo repository.Repository) error {
			// Lookup the secret by ID
			secret, err := repo.Secret().Get(args[0])
			if err != nil {
				return err
			}

			// Verify it's in a signing state
			if secret.Status != "ready" {
				return errors.New("Secret is not in a ready state.")
			}

			cs := NewCombineState(secret.ID, secret.Threshold)

			ioc, _ := lib.Wrap(
				di.ProvideValue(cs),
				di.ProvideValue(sess, di.As(new(ssh.Session))),
			)

			return ioc.Invoke(RunCombineProgram)
		}),
	}

	unsignCmd := &cobra.Command{
		Use:   "unsign {id}",
		Short: "Unsigns a share with a passphrase.",
		Args:  cobra.ExactArgs(1),
		RunE: lib.RunE(func(cmd *cobra.Command, args []string, repo repository.Repository) error {
			// Lookup the secret by ID
			secret, err := repo.Secret().Get(args[0])
			if err != nil {
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
			shares, err := repo.Share().MineForSecret(secret.ID, sess.Context().Value(User{}).(User).ID)
			if err != nil {
				return err
			}

			ioc, err := lib.Wrap(
				di.ProvideValue(cs),
				di.ProvideValue(shares),
				di.ProvideValue(sess, di.As(new(ssh.Session))),
			)

			return ioc.Invoke(RunUnsignProgram)
		}),
	}

	splitCmd.Flags().IntP("parts", "p", 3, "How many shares to split the secret into.")
	splitCmd.Flags().IntP("threshold", "t", 2, "How many shares are required to reconstruct the secret.")

	rootCmd.AddCommand(splitCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(combineCmd)
	rootCmd.AddCommand(unsignCmd)

	return rootCmd
}

package cmd

import (
	"errors"
	"fmt"

	"github.com/adamgoose/ssss/lib"
	"github.com/adamgoose/ssss/lib/model"
	"github.com/adamgoose/ssss/lib/repository"
	"github.com/charmbracelet/ssh"
	"github.com/defval/di"
	"github.com/spf13/cobra"
)

func NewSSHCmd(sess ssh.Session) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sssc",
		Short: "Split and Combine secrets using Shamir's Secret Sharing Scheme.",
		Long: `Split and Combine secrets using Shamir's Secret Sharing Scheme.

Get started by creating an alias in your shell:
  $ alias sssc="ssh -t enge.me --"

Split your first secret:
  $ sssc split
  - Provide a label and your secret
  - Provide the first share encryption passphrase
  - Share the provided "sign" command with your desired shareholders
  - The program exits when all shares are signed

Sign a secret being split:
  $ sssc sign {id}
  - Provide a passphrase to sign the share
  - The program exists after signing

Combine shares to recover a secret:
  $ sssc combine {id}
  - Share the provvided "unsign" command with your shareholders
  - The program exits when all shares are unsigned

Unsign a share:
  $ sssc unsign {id}
  - Provide the passphrase to unsign the share
  - The program exists after unsigning
`,
	}

	lsCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Lists secrets.",
		RunE: lib.RunE(func(cmd *cobra.Command, repo repository.Repository) error {
			out := cmd.OutOrStdout()
			secrets, err := repo.Secret().Mine(sess.Context().Value(model.User{}).(model.User).ID)
			if err != nil {
				return err
			}

			for _, secret := range secrets {
				fmt.Fprintf(out, "%s\t%d/%d\t%s\t%s\n", secret.ID[8:], secret.Threshold, secret.Parts, secret.Label, secret.CreatedAt.Format("2006-01-02 15:04:05"))
			}

			return nil
		}),
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
			shares, err := repo.Share().MineForSecret(secret.ID, sess.Context().Value(model.User{}).(model.User).ID)
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

	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(splitCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(combineCmd)
	rootCmd.AddCommand(unsignCmd)

	return rootCmd
}

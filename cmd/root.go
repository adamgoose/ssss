package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adamgoose/ssss/lib/model"
	"github.com/adamgoose/ssss/lib/repository"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/logging"
)

const (
	host = "0.0.0.0"
	port = "23234"
)

func RunE(repo repository.Repository) error {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return key.Type() == "ssh-ed25519"
		}),
		wish.WithMiddleware(
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					rootCmd := NewSSHCmd(sess)
					rootCmd.SetArgs(sess.Command())
					rootCmd.SetIn(sess)
					rootCmd.SetOut(sess)
					rootCmd.SetErr(sess.Stderr())
					rootCmd.CompletionOptions.DisableDefaultCmd = true
					if err := rootCmd.Execute(); err != nil {
						_ = sess.Exit(1)
						return
					}

					next(sess)
				}
			},
			activeterm.Middleware(),
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					user, err := repo.User().Upsert(&model.User{
						Username:  s.User(),
						PublicKey: base64.StdEncoding.EncodeToString(s.PublicKey().Marshal()),
					})
					if err != nil {
						log.Error("unable to create user", "error", err)
						return
					}

					log.Info("User authenticated", "user.name", user.Username, "user.id", user.ID)
					s.Context().SetValue(model.User{}, *user)

					next(s)
				}
			},
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
	return nil
}

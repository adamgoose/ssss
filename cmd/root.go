package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/logging"
	"github.com/surrealdb/surrealdb.go"
)

const (
	host = "localhost"
	port = "23234"
)

type User struct {
	ID        string `json:"id,omitempty"`
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
}

func RunE(db *surrealdb.DB) error {
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
					data, err := db.Query(`
							INSERT INTO users (id, username, public_key, first_seen, last_seen)
							VALUES ([$username, $public_key], $username, $public_key, time::now(), time::now())
							ON DUPLICATE KEY UPDATE last_seen = time::now()
						`, map[string]interface{}{
						"username":   s.User(),
						"public_key": s.PublicKey().Marshal(),
					})
					if err != nil {
						log.Error("unable to create user", "error", err)
						fmt.Fprintf(s, "unable to create user\n")
						return
					}

					result := make([]struct {
						Result []User `json:"result"`
					}, 1)
					if err := surrealdb.Unmarshal(data, &result); err != nil {
						log.Error("unable to unmarshal user", "error", err)
						return
					}

					user := result[0].Result[0]

					log.Info("User authenticated", "user.name", user.Username, "user.id", user.ID)
					s.Context().SetValue(User{}, user)

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

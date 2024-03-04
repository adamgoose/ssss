package cmd

import (
	"bytes"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/corvus-ch/shamir"
	"github.com/davecgh/go-spew/spew"
	"github.com/surrealdb/surrealdb.go"
)

func NewSplitProgram(s ssh.Session, db *surrealdb.DB) (p *tea.Program) {
	pty, _, ok := s.Pty()

	splitTUI := SplitTUI{
		db:       db,
		term:     pty.Term,
		width:    pty.Window.Width,
		height:   pty.Window.Height,
		session:  s,
		user:     s.Context().Value(User{}).(User),
		renderer: bubbletea.MakeRenderer(s),

		form: huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Key("secret").
					Title("The Secret to Split"),
			).
				WithWidth(pty.Window.Width).
				WithShowHelp(true),
			huh.NewGroup(
				huh.NewInput().
					Key("passphrase").
					Title("Your encryption passphrase"),
				huh.NewInput().
					Key("confirm").
					Title("Confirm your encryption passphrase"),
			).
				WithWidth(pty.Window.Width).
				WithShowHelp(true),
		).
			WithTheme(huh.ThemeCharm()),
	}

	if !ok || s.EmulatedPty() {
		p = tea.NewProgram(splitTUI,
			tea.WithInput(s),
			tea.WithOutput(s),
		)
	} else {
		p = tea.NewProgram(splitTUI,
			tea.WithInput(pty.Slave),
			tea.WithOutput(pty.Slave),
		)
	}

	return
}

type SplitTUI struct {
	db       *surrealdb.DB
	term     string
	width    int
	height   int
	session  ssh.Session
	user     User
	renderer *lipgloss.Renderer

	form       *huh.Form
	secret     *Secret
	splitState *SplitState
}

type Secret struct {
	ID   string `json:"id,omitempty"`
	User string `json:"user"`

	Parts     int       `json:"parts"`
	Threshold int       `json:"threshold"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Share struct {
	ID     string `json:"id,omitempty"`
	Secret string `json:"secret"`
	User   string `json:"user"`

	Key   byte   `json:"key"`
	Share []byte `json:"share"`
}

func (t SplitTUI) Init() tea.Cmd {
	return t.form.Init()
}

func (t SplitTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.height = msg.Height
		t.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if t.secret != nil && t.secret.Status == "signing" {
				t.secret.Status = "dead"
				t.db.Update(t.secret.ID, t.secret)
			}
			return t, tea.Quit
		}
	}

	form, cmd := t.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		t.form = f
	}

	if t.form.State == huh.StateCompleted && t.secret == nil {
		s := Secret{
			User:      t.user.ID,
			Parts:     3,
			Threshold: 2,
			Status:    "signing",
			CreatedAt: time.Now(),
		}

		data, err := t.db.Create("secrets", &s)
		if err != nil {
			return t, tea.Quit
		}

		ns := make([]Secret, 1)
		if err := surrealdb.Unmarshal(data, &ns); err != nil {
			return t, tea.Quit
		}

		t.secret = &ns[0]
		t.splitState = NewSplitState(t.secret.ID, t.secret.Parts)
		t.splitState.Push(Passphrase{
			UserID:     t.user.ID,
			Username:   t.user.Username,
			Passphrase: t.form.GetString("passphrase"),
		})
		t.splitState.Receive()
	}

	if t.splitState != nil && t.splitState.Len() == t.splitState.Expected {
		// Split the secret
		shamirShares, err := shamir.Split([]byte(t.form.GetString("secret")), t.secret.Parts, t.secret.Threshold)
		if err != nil {
			panic(err)
		}

		// Encrypt the Shares
		i := 0
		for k, v := range shamirShares {
			pp := t.splitState.Passphrases[i]
			i++

			cipher, err := encrypt(v, pp.Passphrase)
			if err != nil {
				spew.Dump(err)
				return t, tea.Quit
			}

			share := Share{
				Secret: t.secret.ID,
				User:   t.user.ID,
				Key:    k,
				Share:  cipher,
			}
			spew.Dump(t.db.Create("shares", &share))
		}
		// Store the Shares

		t.secret.Status = "ready"
		t.db.Update(t.secret.ID, t.secret)

		return t, tea.Quit
	}

	return t, cmd
}

func (t SplitTUI) View() string {
	b := bytes.NewBuffer(nil)

	b.WriteString("Split a secret into shares\n\n")
	b.WriteString(fmt.Sprintf("User ID: %s\n\n", t.user.ID))

	if t.secret != nil {
		spew.Fdump(b, t.secret)
		spew.Fdump(b, t.splitState)
	} else {
		b.WriteString(t.form.View())
	}

	return b.String()
}

// import (
// 	"fmt"
//
// 	"github.com/adamgoose/ssss/lib"
// 	"github.com/corvus-ch/shamir"
// 	"github.com/spf13/cobra"
// 	"github.com/surrealdb/surrealdb.go"
// )
//
// var (
// 	parts     int
// 	threshold int
// )
//
// type Secret struct {
// 	ID        string        `json:"id,omitempty"`
// 	Parts     int           `json:"parts"`
// 	Threshold int           `json:"threshold"`
// 	Shares    []SecretShare `json:"shares"`
// }
//
// type SecretShare struct {
// 	ID    string `json:"id,omitempty"`
// 	Label string `json:"label"`
// 	Key   byte   `json:"key"`
// 	Share []byte `json:"share"`
// }
//
// var splitCmd = &cobra.Command{
// 	Use:   "split {secret}",
// 	Short: "Splits a secret into encrypted shares",
// 	RunE: lib.RunE(func(args []string, db *surrealdb.DB) error {
// 		shares := []SecretShare{}
//
// 		rawShares, err := shamir.Split([]byte(args[0]), parts, threshold)
// 		if err != nil {
// 			return err
// 		}
//
// 		for key, rawShare := range rawShares {
// 			shares = append(shares, SecretShare{
// 				Label: fmt.Sprintf("%x", key),
// 				Key:   key,
// 				Share: rawShare,
// 			})
// 		}
//
// 		secret := Secret{
// 			Parts:     parts,
// 			Threshold: threshold,
// 			Shares:    shares,
// 		}
//
// 		data, err := db.Create("secret", secret)
// 		if err != nil {
// 			panic(err)
// 		}
//
// 		storedSecret := make([]Secret, 1)
// 		// storedSecret[0].Shares = make([]SecretShare, parts)
// 		if err := surrealdb.Unmarshal(data, &storedSecret); err != nil {
// 			return err
// 		}
//
// 		fmt.Println(storedSecret[0].ID)
//
// 		return nil
// 	}),
// }
//
// func init() {
// 	rootCmd.AddCommand(splitCmd)
//
// 	splitCmd.Flags().IntVarP(&parts, "parts", "p", 3, "Number of parts to split the secret into")
// 	splitCmd.Flags().IntVarP(&threshold, "threshold", "t", 2, "Number of parts required to reconstruct the secret")
// }

package cmd

import (
	"bytes"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/surrealdb/surrealdb.go"
)

func NewUnsignProgram(s ssh.Session, db *surrealdb.DB, cs *CombineState, shares []Share) (p *tea.Program) {
	pty, _, ok := s.Pty()

	unsignTUI := UnsignTUI{
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
					Key("passphrase").
					Title("Your encryption passphrase"),
			).
				WithWidth(pty.Window.Width).
				WithShowHelp(true),
		).
			WithTheme(huh.ThemeCharm()),

		shares:       shares,
		combineState: cs,
	}

	if !ok || s.EmulatedPty() {
		p = tea.NewProgram(unsignTUI,
			tea.WithInput(s),
			tea.WithOutput(s),
		)
	} else {
		p = tea.NewProgram(unsignTUI,
			tea.WithInput(pty.Slave),
			tea.WithOutput(pty.Slave),
		)
	}

	return
}

type UnsignTUI struct {
	db       *surrealdb.DB
	term     string
	width    int
	height   int
	session  ssh.Session
	user     User
	renderer *lipgloss.Renderer

	form         *huh.Form
	shares       []Share
	combineState *CombineState
}

func (t UnsignTUI) Init() tea.Cmd {
	return t.form.Init()
}

func (t UnsignTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.height = msg.Height
		t.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return t, tea.Quit
		}
	}

	form, cmd := t.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		t.form = f
	}

	if t.form.State == huh.StateCompleted {
		var shamirShare *ShamirShare
		for _, share := range t.shares {
			cipher, err := decrypt(share.Share, t.form.GetString("passphrase"))
			if err != nil {
				continue
			}

			included := false
			for _, ss := range t.combineState.Shares {
				if ss.Key == share.Key {
					included = true
				}
			}
			if included {
				continue
			}

			shamirShare = &ShamirShare{
				Key:   share.Key,
				Share: cipher,
			}
		}

		if shamirShare != nil {
			log.Info("Pushing valid shamir share")
			t.combineState.Push(*shamirShare)
		}

		return t, tea.Quit
	}

	return t, cmd
}

func (t UnsignTUI) View() string {
	b := bytes.NewBuffer(nil)

	b.WriteString("Unsign the secret!\n\n")
	b.WriteString(fmt.Sprintf("User ID: %s\n\n", t.user.ID))

	b.WriteString(t.form.View())

	return b.String()
}

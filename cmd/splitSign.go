package cmd

import (
	"bytes"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/surrealdb/surrealdb.go"
)

func NewSignProgram(s ssh.Session, db *surrealdb.DB, ss *SplitState) (p *tea.Program) {
	pty, _, ok := s.Pty()

	signTUI := SignTUI{
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
				huh.NewInput().
					Key("confirm").
					Title("Confirm your encryption passphrase"),
			).
				WithWidth(pty.Window.Width).
				WithShowHelp(true),
		).
			WithTheme(huh.ThemeCharm()),

		splitState: ss,
	}

	if !ok || s.EmulatedPty() {
		p = tea.NewProgram(signTUI,
			tea.WithInput(s),
			tea.WithOutput(s),
		)
	} else {
		p = tea.NewProgram(signTUI,
			tea.WithInput(pty.Slave),
			tea.WithOutput(pty.Slave),
		)
	}

	return
}

type SignTUI struct {
	db       *surrealdb.DB
	term     string
	width    int
	height   int
	session  ssh.Session
	user     User
	renderer *lipgloss.Renderer

	form *huh.Form
	// secret     *Secret
	splitState *SplitState
}

func (t SignTUI) Init() tea.Cmd {
	return t.form.Init()
}

func (t SignTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		t.splitState.Push(Passphrase{
			UserID:     t.user.ID,
			Username:   t.user.Username,
			Passphrase: t.form.GetString("passphrase"),
		})

		return t, tea.Quit
	}

	return t, cmd
}

func (t SignTUI) View() string {
	b := bytes.NewBuffer(nil)

	b.WriteString("Sign the secret!\n\n")
	b.WriteString(fmt.Sprintf("User ID: %s\n\n", t.user.ID))

	b.WriteString(t.form.View())

	return b.String()
}

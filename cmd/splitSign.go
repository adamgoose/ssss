package cmd

import (
	"github.com/adamgoose/ssss/lib/repository"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
)

func RunSignProgram(s ssh.Session, repo repository.Repository, ss *SplitState) error {
	pty, _, ok := s.Pty()

	signTUI := SignTUI{
		TUI:        NewTUI(s),
		repo:       repo,
		splitState: ss,
	}

	signTUI.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("passphrase").
				Title("Your encryption passphrase"),
			huh.NewInput().
				Key("confirm").
				Title("Confirm your encryption passphrase"),
		),
	).
		WithShowHelp(true).
		WithSubmitCommand(submit).
		WithCancelCommand(cancel).
		WithTheme(huh.ThemeCatppuccin())

	var p *tea.Program
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

	_, err := p.Run()
	return err
}

type SignTUI struct {
	TUI
	repo repository.Repository

	form       *huh.Form
	splitState *SplitState
}

func (t SignTUI) Init() tea.Cmd {
	return t.form.Init()
}

func (t SignTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case submitMsg:
		t.splitState.Push(Passphrase{
			UserID:     t.user.ID,
			Username:   t.user.Username,
			Passphrase: t.form.GetString("passphrase"),
		})

		return t, tea.Quit
	}

	tui, cmd := t.TUI.Update(msg)
	if tui, ok := tui.(TUI); ok {
		t.TUI = tui
		if cmd != nil {
			return t, cmd
		}
	}

	form, cmd := t.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		t.form = f
	}

	return t, cmd
}

func (t SignTUI) View() string {
	v := NewView()

	if t.form.State == huh.StateCompleted {
		v.Colorf(lipgloss.Color("#0F0"), "You signed the secret!")
	} else {
		v.Colorf(lipgloss.Color("#0F0"), "You are signing the secret!")
	}
	v.NL()

	v.WriteString(t.form.View())

	return t.renderer.NewStyle().Width(t.width-2).Border(lipgloss.RoundedBorder(), true).Render(v.String()) + "\n"
}

package cmd

import (
	"github.com/adamgoose/ssss/lib/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

func RunUnsignProgram(s ssh.Session, cs *CombineState, shares []model.Share) error {
	pty, _, ok := s.Pty()

	unsignTUI := UnsignTUI{
		TUI:          NewTUI(s),
		shares:       shares,
		combineState: cs,
	}

	unsignTUI.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("passphrase").
				Title("Your encryption passphrase"),
		),
	).
		WithWidth(pty.Window.Width).
		WithShowHelp(true)

	var p *tea.Program
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

	_, err := p.Run()
	return err
}

type UnsignTUI struct {
	TUI
	form *huh.Form

	shares       []model.Share
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
	v := NewView()

	if t.form.State == huh.StateCompleted {
		v.Colorf(lipgloss.Color("#0F0"), "You unsigned the secret!")
	} else {
		v.Colorf(lipgloss.Color("#0F0"), "You are unsigning the secret!")
	}
	v.NL()

	v.WriteString(t.form.View())

	return t.renderer.NewStyle().Width(t.width-2).Border(lipgloss.RoundedBorder(), true).Render(v.String()) + "\n"
}

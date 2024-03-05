package cmd

import (
	"github.com/adamgoose/ssss/lib/repository"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/corvus-ch/shamir"
)

func RunCombineProgram(s ssh.Session, repo repository.Repository, cs *CombineState) error {
	pty, _, ok := s.Pty()

	combineTUI := CombineTUI{
		TUI:          NewTUI(s),
		repo:         repo,
		progress:     progress.New(progress.WithWidth(pty.Window.Width-2), progress.WithoutPercentage()),
		combineState: cs,
	}

	var p *tea.Program
	if !ok || s.EmulatedPty() {
		p = tea.NewProgram(combineTUI,
			tea.WithInput(s),
			tea.WithOutput(s),
		)
	} else {
		p = tea.NewProgram(combineTUI,
			tea.WithInput(pty.Slave),
			tea.WithOutput(pty.Slave),
		)
	}

	_, err := p.Run()
	return err
}

type CombineTUI struct {
	TUI
	repo     repository.Repository
	progress progress.Model

	combineState *CombineState
	secret       *[]byte
}

func (t CombineTUI) Init() tea.Cmd {
	return tea.Batch(
		receive(t.combineState),
		t.TUI.Init(),
	)
}

func (t CombineTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case receiveMsg:
		log.Info("Received a share")

		// Have I received sufficient shares?
		if t.combineState.Len() == t.combineState.Expected {
			log.Info("Received all shares")
			return t, receivedAll
		}

		// Wait for another one
		return t, receive(t.combineState)
	case receivedAllMsg:
		shares := map[byte][]byte{}
		for _, share := range t.combineState.Shares {
			shares[share.Key] = share.Share
		}

		v, err := shamir.Combine(shares)
		if err == nil {
			t.secret = &v
		}

		return t, tea.Quit
	}

	tui, cmd := t.TUI.Update(msg)
	if tui, ok := tui.(TUI); ok {
		t.TUI = tui
		if cmd != nil {
			return t, cmd
		}
	}

	return t, nil
}

func (t CombineTUI) View() string {
	v := NewView()

	if t.secret != nil {
		v.WriteString("Your secret is: ")
		v.Colorf(lipgloss.Color("#0F0"), string(*t.secret))
	} else {
		v.WriteString("Ask others to unsign their shares with: ")
		v.Colorf(lipgloss.Color("#0F0"), "ssh -t enge.me -- unsign %s", t.combineState.SecretID[8:])
		v.NL()
		v.WriteString(t.progress.ViewAs(float64(t.combineState.Len()) / float64(t.combineState.Expected)))
	}

	return t.renderer.NewStyle().Width(t.width-2).Border(lipgloss.RoundedBorder(), true).Render(v.String()) + "\n"
}

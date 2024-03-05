package cmd

import (
	"github.com/adamgoose/ssss/lib/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
)

type TUI struct {
	term     string
	width    int
	height   int
	session  ssh.Session
	user     model.User
	renderer *lipgloss.Renderer
}

func NewTUI(s ssh.Session) TUI {
	pty, _, _ := s.Pty()

	return TUI{
		term:     pty.Term,
		width:    pty.Window.Width,
		height:   pty.Window.Height,
		session:  s,
		user:     s.Context().Value(model.User{}).(model.User),
		renderer: bubbletea.MakeRenderer(s),
	}
}

func (t TUI) Init() tea.Cmd {
	return nil
}

func (t TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	return t, nil
}

func (t TUI) View() string {
	return ""
}

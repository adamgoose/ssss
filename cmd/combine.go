package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/corvus-ch/shamir"
	"github.com/davecgh/go-spew/spew"
	"github.com/surrealdb/surrealdb.go"
)

func NewCombineProgram(s ssh.Session, db *surrealdb.DB, cs *CombineState) (p *tea.Program) {
	pty, _, ok := s.Pty()

	combineTUI := CombineTUI{
		db:           db,
		term:         pty.Term,
		width:        pty.Window.Width,
		height:       pty.Window.Height,
		session:      s,
		user:         s.Context().Value(User{}).(User),
		renderer:     bubbletea.MakeRenderer(s),
		combineState: cs,
	}
	cs.Receive()

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

	return p
}

type CombineTUI struct {
	db           *surrealdb.DB
	term         string
	width        int
	height       int
	session      ssh.Session
	user         User
	renderer     *lipgloss.Renderer
	combineState *CombineState
}

func (t CombineTUI) Init() tea.Cmd {
	return nil
}

func (t CombineTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	if t.combineState.Len() == t.combineState.Expected {
		shares := map[byte][]byte{}
		for _, share := range t.combineState.Shares {
			shares[share.Key] = share.Share
		}

		spew.Dump(shamir.Combine(shares))
		return t, tea.Quit
	}

	return t, nil
}

func (t CombineTUI) View() string {
	return spew.Sdump(t.combineState)
}

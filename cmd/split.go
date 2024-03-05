package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/adamgoose/ssss/lib/model"
	"github.com/adamgoose/ssss/lib/repository"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/corvus-ch/shamir"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
)

type (
	submitMsg      struct{}
	cancelMsg      struct{}
	receiveMsg     struct{}
	receivedAllMsg struct{}
)

func submit() tea.Msg {
	return submitMsg{}
}

func cancel() tea.Msg {
	return cancelMsg{}
}

func receive(state interface {
	ReceiveOne() error
},
) tea.Cmd {
	return func() tea.Msg {
		log.Info("Receiving a passphrase")
		state.ReceiveOne()
		return receiveMsg{}
	}
}

func receivedAll() tea.Msg {
	return receivedAllMsg{}
}

func RunSplitProgram(s ssh.Session, repo repository.Repository, cmd *cobra.Command) error {
	pty, _, ok := s.Pty()

	parts, _ := cmd.Flags().GetInt("parts")
	threshold, _ := cmd.Flags().GetInt("threshold")

	splitTUI := SplitTUI{
		TUI:       NewTUI(s),
		repo:      repo,
		progress:  progress.New(progress.WithWidth(pty.Window.Width-2), progress.WithoutPercentage()),
		parts:     parts,
		threshold: threshold,
	}

	splitTUI.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("label").
				Title("Label").
				Description("An insecure label for your secret"),
			huh.NewText().
				Key("secret").
				Title("Secret").
				Description("The secret you want to split"),
		),
		huh.NewGroup(
			huh.NewInput().
				Key("passphrase").
				Title("Passphrase").
				Description("A secure passphrase to encrypt your secret"),
			huh.NewInput().
				Key("confirm").
				Title("Confirm Passphrase").
				Description("Confirm your secure passphrase"),
		),
	).
		WithShowHelp(true).
		WithSubmitCommand(submit).
		WithCancelCommand(cancel).
		WithTheme(huh.ThemeCatppuccin())

	var p *tea.Program
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

	_, err := p.Run()
	return err
}

type SplitTUI struct {
	TUI
	repo     repository.Repository
	form     *huh.Form
	progress progress.Model

	parts     int
	threshold int

	secret     *model.Secret
	splitState *SplitState
}

func (t SplitTUI) Init() tea.Cmd {
	return tea.Batch(
		t.TUI.Init(),
		t.form.Init(),
	)
}

func (t SplitTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case submitMsg:
		s, err := t.repo.Secret().Create(&model.Secret{
			User:      t.user.ID,
			Label:     t.form.GetString("label"),
			Parts:     t.parts,
			Threshold: t.threshold,
			Status:    "signing",
			CreatedAt: time.Now(),
		})
		if err != nil {
			return t, tea.Quit
		}

		log.Info("Splitting a Secret", "id", s.ID, "user", t.user.ID)

		t.secret = s
		t.splitState = NewSplitState(t.secret.ID, t.secret.Parts)
		t.splitState.Push(Passphrase{
			UserID:     t.user.ID,
			Username:   t.user.Username,
			Passphrase: t.form.GetString("passphrase"),
		})

		return t, receive(t.splitState)
	case receiveMsg:
		log.Info("Received a passphrase")

		// Have I received all the passphrases?
		if t.splitState.Len() == t.splitState.Expected {
			return t, receivedAll
		}

		// Wait for another one
		return t, receive(t.splitState)
	case receivedAllMsg:

		log.Info("Received all message")

		// Split the secret
		shamirShares, err := shamir.Split([]byte(t.form.GetString("secret")), t.secret.Parts, t.secret.Threshold)
		if err != nil {
			panic(err)
		}

		// Encrypt and store the Shares
		i := 0
		for k, v := range shamirShares {
			pp := t.splitState.Passphrases[i]
			i++

			cipher, err := encrypt(v, pp.Passphrase)
			if err != nil {
				spew.Dump(err)
				return t, tea.Quit
			}

			t.repo.Share().Create(&model.Share{
				Secret: t.secret.ID,
				User:   pp.UserID,
				Key:    k,
				Share:  cipher,
			})
		}

		delete(SplitStates, t.secret.ID)
		t.secret.Status = "ready"
		t.repo.Secret().Update(t.secret)

		return t, tea.Quit
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if t.secret != nil && t.secret.Status == "signing" {
				delete(SplitStates, t.secret.ID)
				t.secret.Status = "dead"
				t.repo.Secret().Update(t.secret)
			}
		}
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
		if cmd != nil {
			return t, cmd
		}
	}

	prog, cmd := t.progress.Update(msg)
	if p, ok := prog.(progress.Model); ok {
		t.progress = p
		if cmd != nil {
			return t, cmd
		}
	}

	return t, nil
}

type View struct {
	*strings.Builder
}

func NewView() *View {
	return &View{&strings.Builder{}}
}

func (v View) Colorf(color lipgloss.TerminalColor, format string, a ...interface{}) {
	v.WriteString(lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf(format, a...)))
}

func (v View) NL() {
	v.WriteRune('\n')
}

func (t SplitTUI) View() string {
	v := NewView()

	if t.secret != nil {
		v.Colorf(lipgloss.Color("#0F0"), "Status: %s", t.secret.Status)
		v.NL()

		if t.secret.Status == "signing" {
			v.WriteString("Ask others to sign their shares with: ")
			v.Colorf(lipgloss.Color("#0F0"), "ssh -t enge.me -- sign %s", t.secret.ID[8:])
			v.NL()
			v.WriteString(t.progress.ViewAs(float64(t.splitState.Len()) / float64(t.splitState.Expected)))
		}
		if t.secret.Status == "ready" {
			v.WriteString("Retrieve your secret with: ")
			v.Colorf(lipgloss.Color("#0F0"), "ssh -t enge.me -- combine %s", t.secret.ID[8:])
			v.NL()
		}

		for _, pass := range t.splitState.Passphrases {
			v.NL()
			v.Colorf(lipgloss.Color("#AFA"), "User: %s ", pass.Username)
			v.Colorf(lipgloss.Color("#FAA"), "Length: %d", len(pass.Passphrase))
		}

	} else {
		v.WriteString(
			t.form.View(),
		)
	}

	return t.renderer.NewStyle().Width(t.width-2).Border(lipgloss.RoundedBorder(), true).Render(v.String()) + "\n"
}

package cmd

import "sync"

var SplitStates = make(map[string]*SplitState)

func NewSplitState(secretId string, expected int) *SplitState {
	s := &SplitState{
		SecretID:       secretId,
		Expected:       expected,
		Passphrases:    make([]Passphrase, 0),
		chanPassphrase: make(chan Passphrase, expected),
		chanDone:       make(chan error, 1),
	}

	SplitStates[secretId] = s
	return s
}

type SplitState struct {
	mu             sync.Mutex
	chanPassphrase chan Passphrase
	chanDone       chan error

	SecretID    string
	Expected    int
	Passphrases []Passphrase
}

type Passphrase struct {
	UserID     string
	Username   string
	Passphrase string
}

func (s *SplitState) Len() int {
	return len(s.Passphrases)
}

func (s *SplitState) Push(p Passphrase) {
	s.chanPassphrase <- p
}

func (s *SplitState) Receive() {
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		for i := 0; i < s.Expected; i++ {
			p := <-s.chanPassphrase
			s.Passphrases = append(s.Passphrases, p)
		}

		s.chanDone <- nil
	}()
}

func (s *SplitState) ReceiveOne() error {
	p := <-s.chanPassphrase

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Passphrases = append(s.Passphrases, p)

	if len(s.Passphrases) == s.Expected {
		s.chanDone <- nil
	}

	return nil
}

package cmd

import "sync"

var SplitStates = make(map[string]*SplitState)

func NewSplitState(secretId string, expected int) *SplitState {
	s := &SplitState{
		SecretID:    secretId,
		Expected:    expected,
		Passphrases: make([]Passphrase, 0),
		chanP:       make(chan Passphrase, expected),
	}

	SplitStates[secretId] = s
	return s
}

type SplitState struct {
	mu       sync.Mutex
	chanP    chan Passphrase
	chanDone chan error

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
	s.chanP <- p
}

func (s *SplitState) Receive() {
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		for i := 0; i < s.Expected; i++ {
			p := <-s.chanP
			s.Passphrases = append(s.Passphrases, p)
		}

		s.chanDone <- nil
	}()
}

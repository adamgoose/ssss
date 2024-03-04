package cmd

import "sync"

var CombineStates = make(map[string]*CombineState)

type ShamirShare struct {
	Key   byte
	Share []byte
}

func NewCombineState(secretId string, expected int) *CombineState {
	s := &CombineState{
		SecretID: secretId,
		Expected: expected,
		Shares:   make([]ShamirShare, 0),
		chanS:    make(chan ShamirShare, expected),
	}

	CombineStates[secretId] = s
	return s
}

type CombineState struct {
	mu       sync.Mutex
	chanS    chan ShamirShare
	chanDone chan error

	SecretID string
	Expected int
	Shares   []ShamirShare
}

func (c *CombineState) Len() int {
	return len(c.Shares)
}

func (c *CombineState) Push(s ShamirShare) {
	c.chanS <- s
}

func (c *CombineState) Receive() {
	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		for i := 0; i < c.Expected; i++ {
			s := <-c.chanS
			c.Shares = append(c.Shares, s)
		}

		c.chanDone <- nil
	}()
}

package model

type Share struct {
	ID     string `json:"id,omitempty"`
	Secret string `json:"secret"`
	User   string `json:"user"`

	Key   byte   `json:"key"`
	Share []byte `json:"share"`
}

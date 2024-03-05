package model

import "time"

type Secret struct {
	ID   string `json:"id,omitempty"`
	User string `json:"user"`

	Label     string    `json:"label"`
	Parts     int       `json:"parts"`
	Threshold int       `json:"threshold"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

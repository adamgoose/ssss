package model

type User struct {
	ID        string `json:"id,omitempty"`
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
	// FirstSeen time.Time `json:"first_seen"`
	// LastSeen  time.Time `json:"last_seen"`
}

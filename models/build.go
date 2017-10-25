package models

type Build struct {
	ID     string
	Number int
	Status string
	Link   string

	// Project
	Owner string
	Repo  string

	// Git specific
	CloneURL string
	Commit   string
	Remote   string
}

package model

type Image struct {
	ID        int        `json:"id"`
	Apartment *Apartment `json:"apartment"`
	Address   string     `json:"address"`
}

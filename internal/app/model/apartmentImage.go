package model

type ApartmentImage struct {
	ID        int        `json:"id"`
	Apartment *Apartment `json:"apartment"`
	Address   string     `json:"address"`
}

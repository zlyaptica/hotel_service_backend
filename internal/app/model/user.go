package model

import validation "github.com/go-ozzo/ozzo-validation"

type User struct {
	ID          int    `json:"id"`
	LName       string `json:"l_name"`
	FName       string `json:"f_name"`
	PhoneNumber string `json:"phone_number"`
}

func (g *User) Validate() error {
	return validation.ValidateStruct(
		g,
		validation.Field(&g.PhoneNumber, validation.Required, validation.Length(11, 14)),
		validation.Field(&g.LName, validation.Required, validation.Length(3, 20)),
	)
}

package model

import validation "github.com/go-ozzo/ozzo-validation"

type Address struct {
	ID      int    `json:"id"`
	Country string `json:"country"`
	City    string `json:"city"`
	Street  string `json:"street"`
	House   string `json:"house"`
}

func (a *Address) Validate() error {
	return validation.ValidateStruct(
		a,
		validation.Field(&a.Country, validation.Required, validation.Length(5, 40)),
		validation.Field(&a.City, validation.Required, validation.Length(5, 40)),
		validation.Field(&a.Street, validation.Required, validation.Length(5, 40)),
		validation.Field(&a.House, validation.Required, validation.Length(1, 10)),
	)
}

package model

import validation "github.com/go-ozzo/ozzo-validation"

type Apartment struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	Hotel          *Hotel          `json:"hotel"`
	ApartmentClass *ApartmentClass `json:"apartment_class"`
	IsFree         *bool           `json:"is_free"`
	BedCount       int             `json:"bed_count"`
	Price          int             `json:"price"`
}

func (a *Apartment) Validate() error {
	return validation.ValidateStruct(
		a,
		validation.Field(&a.Name, validation.Required, validation.Length(10, 40)),
		validation.Field(&a.BedCount, validation.Required, validation.Length(1, 4)),
		validation.Field(&a.Price, validation.Required, validation.Length(1000, 9999999)),
	)
}

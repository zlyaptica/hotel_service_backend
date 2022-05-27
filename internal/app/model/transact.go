package model

import "time"

type Transact struct {
	ID            int        `json:"id"`
	OperationDate time.Time  `json:"operation_date"`
	Apartment     *Apartment `json:"apartment"`
	User          *User      `json:"user"`
	Price         int        `json:"price"`
	DateArrival   time.Time  `json:"date_arrival"`
	DateDeparture time.Time  `json:"date_departure"`
}

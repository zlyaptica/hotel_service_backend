package sqlstore

import (
	"database/sql"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
	"github.com/zlyaptica/hotel_service_backend/store"
	"time"
)

type TransactRepository struct {
	store *Store
}

func (r TransactRepository) Create(t *model.Transact) error {
	q := `INSERT INTO transact (apartment_id, guest_id, date_arrival, date_departure, price, date) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return r.store.db.QueryRow(
		q,
		t.Apartment.ID,
		t.User.ID,
		t.DateArrival,
		t.DateDeparture,
		t.Price,
		time.Now(),
	).Scan(&t.ID)
}

func (r TransactRepository) CreateTransact(t *model.Transact) error {
	q := `INSERT INTO transact (apartment_id, guest_id, date_arrival, date_departure, price, date) 
		  VALUES ($1, (SELECT id FROM guests WHERE phone_number = $2), $3, $4, $5, $6) RETURNING id`
	return r.store.db.QueryRow(
		q,
		t.Apartment.ID,
		t.User.PhoneNumber,
		t.DateArrival,
		t.DateDeparture,
		t.Price,
		time.Now(),
	).Scan(&t.ID)
}

func (r TransactRepository) FindTransactsByPhoneNumber(phoneNumber string) ([]model.Transact, error) {
	transacts := []model.Transact{}
	q := `SELECT t.id, g.id, g.phone_number, t.price, t.date, t.date_arrival, t.date_departure, 
       a.id, a.bed_count, a.is_free, a.name, a.price, ac.id, ac.class, h.id, h.name
       FROM transact t
			INNER JOIN guests g on t.guest_id = g.id
			INNER JOIN apartments a on a.id = t.apartment_id
       		INNER JOIN apartment_classes ac on ac.id = a.apartment_class_id
       		INNER JOIN hotels h on h.id = a.hotel_id
			WHERE g.phone_number = $1`
	rows, err := r.store.db.Query(q, phoneNumber)
	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}
	for rows.Next() {
		u := &model.User{}
		ac := &model.ApartmentClass{}
		h := &model.Hotel{}
		a := &model.Apartment{
			Hotel:          h,
			ApartmentClass: ac,
		}
		t := model.Transact{
			User:      u,
			Apartment: a,
		}
		err := rows.Scan(
			&t.ID,
			&t.User.ID,
			&t.User.PhoneNumber,
			&t.Price,
			&t.OperationDate,
			&t.DateArrival,
			&t.DateDeparture,
			&t.Apartment.ID,
			&t.Apartment.BedCount,
			&t.Apartment.IsFree,
			&t.Apartment.Name,
			&t.Apartment.Price,
			&t.Apartment.ApartmentClass.ID,
			&t.Apartment.ApartmentClass.Class,
			&t.Apartment.Hotel.ID,
			&t.Apartment.Hotel.Name,
		)
		if err != nil {
			return nil, err
		}
		transacts = append(transacts, t)
	}
	return transacts, nil
}

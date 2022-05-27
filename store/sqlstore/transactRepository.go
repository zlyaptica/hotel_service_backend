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

func (r TransactRepository) FindTransactsByUserID(id int) ([]model.Transact, error) {
	transacts := []model.Transact{}
	q := `SELECT t.id, a.name, h.name, t.date_arrival, t.date_departure, t.date, t.price FROM transact t
		  INNER JOIN guests g on g.id = t.guest_id
		  INNER JOIN apartments a on t.apartment_id = a.id
		  INNER JOIN hotels h on a.hotel_id = h.id
          WHERE g.id = $1`
	rows, err := r.store.db.Query(q, id)
	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}
	for rows.Next() {
		u := &model.User{}
		a := &model.Apartment{}
		t := model.Transact{
			User:      u,
			Apartment: a,
		}
		err := rows.Scan(
			&t.ID,
			&t.Apartment.Name,
			&t.Apartment.Hotel.ID,
			&t.DateArrival,
			t.DateDeparture,
			t.OperationDate,
			t.Price,
		)
		if err != nil {
			return nil, err
		}
	}
	return transacts, nil
}

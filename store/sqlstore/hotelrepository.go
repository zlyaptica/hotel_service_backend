package sqlstore

import (
	"database/sql"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
	"github.com/zlyaptica/hotel_service_backend/store"
)

type HotelRepository struct {
	store *Store
}

func (r HotelRepository) FindAll() ([]model.Hotel, error) {
	hotels := []model.Hotel{}

	q := `SELECT h.id, a.id, h.name, h.stars_count, a.country, a.city, a.street, a.house 
		  FROM hotels h
		  INNER JOIN address a on h.address_id = a.id`
	rows, err := r.store.db.Query(q)
	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	for rows.Next() {
		a := &model.Address{}
		h := model.Hotel{
			Address: a,
		}
		err := rows.Scan(
			&h.ID,
			&h.Address.ID,
			&h.Name,
			&h.StarsCount,
			&h.Address.Country,
			&h.Address.City,
			&h.Address.Street,
			&h.Address.House,
		)
		if err != nil {
			return nil, err
		}
		hotels = append(hotels, h)
	}
	return hotels, nil
}

func (r HotelRepository) Find(id int) (*model.Hotel, error) {
	a := &model.Address{}
	h := &model.Hotel{
		Address: a,
	}
	q := `SELECT h.id, a.id, h.name, h.stars_count, a.country, a.city, a.street, a.house 
		  FROM hotels h
		  INNER JOIN address a on h.address_id = a.id
		  WHERE h.id = $1`
	if err := r.store.db.QueryRow(
		q,
		id,
	).Scan(
		&h.ID,
		&h.Address.ID,
		&h.Name,
		&h.StarsCount,
		&h.Address.Country,
		&h.Address.City,
		&h.Address.Street,
		&h.Address.House,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}
	return h, nil
}

func (r HotelRepository) FindByCountry(country string) ([]model.Hotel, error) {
	hotels := []model.Hotel{}
	q := `SELECT h.id, a.id, h.name, h.stars_count, a.country, a.city, a.street, a.house 
		  FROM hotels h
		  INNER JOIN address a on h.address_id = a.id
		  WHERE a.country = $1`
	rows, err := r.store.db.Query(q, country)
	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	for rows.Next() {
		a := &model.Address{}
		h := model.Hotel{
			Address: a,
		}
		err := rows.Scan(
			&h.ID,
			&h.Address.ID,
			&h.Name,
			&h.StarsCount,
			&h.Address.Country,
			&h.Address.City,
			&h.Address.Street,
			&h.Address.House,
		)
		if err != nil {
			return nil, err
		}

		hotels = append(hotels, h)
	}

	return hotels, nil
}

func (r HotelRepository) FindByCity(city string) ([]model.Hotel, error) {
	hotels := []model.Hotel{}
	q := `SELECT h.id, a.id, h.name, h.stars_count, a.country, a.city, a.street, a.house 
		  FROM hotels h
		  INNER JOIN address a on h.address_id = a.id
		  WHERE a.city = $1`
	rows, err := r.store.db.Query(q, city)
	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}
	for rows.Next() {
		a := &model.Address{}
		h := model.Hotel{
			Address: a,
		}
		err := rows.Scan(
			&h.ID,
			&h.Address.ID,
			&h.Name,
			&h.StarsCount,
			&h.Address.Country,
			&h.Address.City,
			&h.Address.Street,
			&h.Address.House,
		)

		if err != nil {
			return nil, err
		}
	}
	return hotels, nil
}

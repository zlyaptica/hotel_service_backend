package sqlstore

import (
	"database/sql"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
	"github.com/zlyaptica/hotel_service_backend/store"
)

type HotelRepository struct {
	store *Store
}

func (r HotelRepository) Create(hotel *model.Hotel) error {
	q := `INSERT INTO address (country, city, street, house) VALUES ($1, $2, $3, $4) RETURNING id`
	var addressID int
	_ = r.store.db.QueryRow(q,
		hotel.Address.Country,
		hotel.Address.City,
		hotel.Address.Street,
		hotel.Address.House,
	).Scan(&addressID)
	q = `INSERT INTO hotels (name, address_id, stars_count, description, header_image_address) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return r.store.db.QueryRow(
		q,
		hotel.Name,
		addressID,
		hotel.StarsCount,
		hotel.Description,
		hotel.HeaderImageAddress,
	).Scan(&hotel.ID)
}

func (r HotelRepository) Delete(id int) error {
	q := `SELECT address_id FROM hotels where id = $1`
	var addressID int
	_ = r.store.db.QueryRow(q, id).Scan(&addressID)

	q = `DELETE FROM hotels WHERE id = $1`
	_, err := r.store.db.Query(q, id)

	q = `DELETE FROM address WHERE id = $1`
	_, err = r.store.db.Query(q, addressID)
	return err
}
func (r HotelRepository) Update(hotel *model.Hotel) error {
	q := `SELECT address_id FROM hotels WHERE id = $1`
	var addressID int
	_ = r.store.db.QueryRow(q, hotel.ID).Scan(&addressID)

	q = `UPDATE address SET (country, city, street, house) = ($1, $2, $3, $4) WHERE id = $5`
	_, err := r.store.db.Query(
		q,
		hotel.Address.Country,
		hotel.Address.City,
		hotel.Address.Street,
		hotel.Address.House,
		addressID,
	)

	q = `UPDATE hotels SET (name, stars_count, description, header_image_address) = ($1, $2, $3, $4) WHERE id = $5`
	_, err = r.store.db.Query(
		q,
		hotel.Name,
		hotel.StarsCount,
		hotel.Description,
		hotel.HeaderImageAddress,
		hotel.ID,
	)
	return err
}

func (r HotelRepository) FindAll() ([]model.Hotel, error) {
	hotels := []model.Hotel{}

	q := `SELECT h.id, a.id, h.name, h.description, h.header_image_address, h.stars_count, a.country, a.city, a.street, a.house 
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
			&h.Description,
			&h.HeaderImageAddress,
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
	q := `SELECT h.id, a.id, h.name, h.description, h.header_image_address, h.stars_count, a.country, a.city, a.street, a.house 
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
		&h.Description,
		&h.HeaderImageAddress,
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
	q := `SELECT h.id, a.id, h.name, h.description, h.header_image_address, h.stars_count, a.country, a.city, a.street, a.house 
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
			&h.Description,
			&h.HeaderImageAddress,
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
	q := `SELECT h.id, a.id, h.name, h.description, h.header_image_address, h.stars_count, a.country, a.city, a.street, a.house 
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
			&h.Description,
			&h.HeaderImageAddress,
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

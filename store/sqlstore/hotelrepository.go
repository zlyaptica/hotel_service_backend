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
	hotels := []model.Hotel{} // массив структур

	q := `SELECT h.id, a.id, h.name, h.description, h.header_image_address, h.stars_count, a.country, a.city, a.street, a.house 
		  FROM hotels h
		  INNER JOIN address a on h.address_id = a.id`
	rows, err := r.store.db.Query(q) // in rows заносим строки с помощью пакета database/sql, в ерр ошибку
	defer rows.Close()               // закрываем бд(закроется при выходе из функции)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound // если ничего не найдено, то возвращаем ничего для массива и
			// ошибку о том, что ничего не найдено
		}
		return nil, err // в ином случае возвращаем ничего и полученную ошибку
	}

	for rows.Next() { // для каждого элемета массива:
		a := &model.Address{} // а присваиваем ссылку на структуру с моделью адреса
		h := model.Hotel{
			Address: a, // в качестве адреса берем ссылку на адрес
		}
		err := rows.Scan( // из запросы сканируем выбранные поля в структуру
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
		} // если есть ошибка, то возвращаем пустой слайс отелей и ошибку
		hotels = append(hotels, h) // если все ок, то добавляем отель в слайс
	}
	return hotels, nil // возвращаем слайс отелей и ноль для ошибки
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

package sqlstore

import (
	"database/sql"
	"encoding/json"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
	"github.com/zlyaptica/hotel_service_backend/store"
)

type ApartmentRepository struct {
	store *Store
}

func (r ApartmentRepository) Create(bedCount, price, apartmentClassID, hotelID json.Number, name string) error {
	a := &model.Apartment{}
	q := `INSERT INTO apartments (hotel_id, is_free, bed_count, price, apartment_class_id, name) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	return r.store.db.QueryRow(
		q,
		hotelID,
		true,
		bedCount,
		price,
		apartmentClassID,
		name,
	).Scan(&a.ID)
}

func (r ApartmentRepository) FindAll() ([]model.Apartment, error) {
	apartments := []model.Apartment{}
	q := `SELECT a.id, h.id, adr.id, ac.id, a.is_free, a.bed_count, a.price, ac.class, h.name, 
                 h.stars_count, adr.country, adr.city, adr.street, adr.house
		  FROM apartments a
          INNER JOIN hotels h ON a.hotel_id = h.id
          INNER JOIN address adr ON h.address_id = adr.id
          INNER JOIN apartment_classes ac ON a.apartment_class_id = ac.id`
	rows, err := r.store.db.Query(q)
	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	for rows.Next() {
		ac := &model.ApartmentClass{}
		adr := &model.Address{}
		h := &model.Hotel{
			Address: adr,
		}
		a := model.Apartment{
			Hotel:          h,
			ApartmentClass: ac,
		}
		err := rows.Scan(
			&a.ID,
			&a.Hotel.ID,
			&a.Hotel.Address.ID,
			&a.ApartmentClass.ID,
			&a.IsFree,
			&a.BedCount,
			&a.Price,
			&a.ApartmentClass.Class,
			&a.Hotel.Name,
			&a.Hotel.StarsCount,
			&a.Hotel.Address.Country,
			&a.Hotel.Address.City,
			&a.Hotel.Address.Street,
			&a.Hotel.Address.House,
		)

		if err != nil {
			return nil, err
		}
		apartments = append(apartments, a)
	}
	return apartments, nil
}

func (r ApartmentRepository) GetPriceApartment(id int) (int, error) {
	var price int
	q := `SELECT price FROM apartments WHERE id = $1`
	if err := r.store.db.QueryRow(q, id).Scan(
		&price,
	); err != nil {
		return 0, err
	}
	return price, nil
}

func (r ApartmentRepository) FillRoom(id int) error {
	q := `UPDATE apartments SET is_free = false WHERE id = $1`
	_, err := r.store.db.Exec(q, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return store.ErrRecordNotFound
		}
		return err
	}
	return nil
}

func (r ApartmentRepository) FindByHotelID(id int) ([]model.Apartment, error) {
	apartments := []model.Apartment{}
	q := `SELECT a.id, a.hotel_id, a.is_free, a.bed_count, a.price, ac.class, a.name FROM apartments a
			INNER JOIN apartment_classes ac on ac.id = a.apartment_class_id
			WHERE hotel_id = $1 AND is_free = true`
	rows, err := r.store.db.Query(q, id)
	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	for rows.Next() {
		ac := &model.ApartmentClass{}
		h := &model.Hotel{}
		a := model.Apartment{
			Hotel:          h,
			ApartmentClass: ac,
		}
		err := rows.Scan(
			&a.ID,
			&a.Hotel.ID,
			&a.IsFree,
			&a.BedCount,
			&a.Price,
			&a.ApartmentClass.Class,
			&a.Name,
		)
		if err != nil {
			return nil, err
		}
		apartments = append(apartments, a)
	}
	return apartments, nil
}

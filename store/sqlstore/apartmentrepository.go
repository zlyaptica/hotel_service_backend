package sqlstore

import (
	"database/sql"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
	"github.com/zlyaptica/hotel_service_backend/store"
)

type ApartmentRepository struct {
	store *Store
}

func (r ApartmentRepository) Create(bedCount, price, apartmentClassID, hotelID int) error {
	//ac := &model.ApartmentClass{
	//	ID: apartmentClassID,
	//}
	//h := &model.Hotel{
	//	ID: hotelID,
	//}
	//a := &model.Apartment{
	//	BedCount:       bedCount,
	//	Price:          price,
	//	Hotel:          h,
	//	ApartmentClass: ac,
	//}
	//if err := a.Validate(); err != nil {
	//	return err
	//}
	a := &model.Apartment{}
	q := `INSERT INTO apartments (hotel_id, is_free, bed_count, price, apartment_class_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	return r.store.db.QueryRow(
		q,
		hotelID,
		true,
		bedCount,
		price,
		apartmentClassID,
	).Scan(&a.ID)
}

func (r ApartmentRepository) Delete(id int) error {
	q := `DELETE FROM apartments WHERE id = $1`
	_, err := r.store.db.Query(q, id)
	return err
}

func (r ApartmentRepository) FindAll() ([]model.Apartment, error) {
	apartments := []model.Apartment{}
	q := `SELECT a.id, h.id, adr.id, ac.id, a.is_free, a.bed_count, a.price, ac.class, h.name, 
                 h.stars_count, adr.country, adr.city, adr.street, adr.house
		  FROM apartments a
          INNER JOIN hotels h ON a.hotel_id = h.id
          INNER JOIN address adr ON h.address_id = adr.id
          INNER JOIN apartments_classes ac ON a.apartment_class_id = ac.id`
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

func (r ApartmentRepository) Find(id int) (*model.Apartment, error) {
	ac := &model.ApartmentClass{}
	adr := &model.Address{}
	h := &model.Hotel{
		Address: adr,
	}
	a := &model.Apartment{
		Hotel:          h,
		ApartmentClass: ac,
	}
	q := `SELECT a.id, h.id, adr.id, ac.id, a.is_free, a.bed_count, a.price, ac.class, h.name, 
                 h.stars_count, adr.country, adr.city, adr.street, adr.house
		  FROM apartments a
          INNER JOIN hotels h ON a.hotel_id = h.id
          INNER JOIN address adr ON h.address_id = adr.id
          INNER JOIN apartments_classes ac ON a.apartment_class_id = ac.id
          WHERE a.id = $1`
	if err := r.store.db.QueryRow(q, id).Scan(
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
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}
	return a, nil
}

func (r ApartmentRepository) FindByHotel(hotel string) ([]model.Apartment, error) {
	//TODO implement me
	panic("implement me")

}

func (r ApartmentRepository) FindByBedCount(count int) ([]model.Apartment, error) {
	//TODO implement me
	panic("implement me")
}

func (r ApartmentRepository) FindFree(bool) ([]model.Apartment, error) {
	//TODO implement me
	panic("implement me")
}

func (r ApartmentRepository) FindByClass(class string) ([]model.Apartment, error) {
	//TODO implement me
	panic("implement me")
}

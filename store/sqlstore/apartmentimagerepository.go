package sqlstore

import (
	"database/sql"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
	"github.com/zlyaptica/hotel_service_backend/store"
)

type ApartmentImageRepository struct {
	store *Store
}

func (r ApartmentImageRepository) GetImagesByHotelID(id int) ([]model.ApartmentImage, error) {
	images := []model.ApartmentImage{}
	q := `SELECT id, hotel_id, address FROM apartment_images WHERE hotel_id = $1`
	rows, err := r.store.db.Query(q, id)
	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	for rows.Next() {
		a := &model.Apartment{}
		i := model.ApartmentImage{
			Apartment: a,
		}
		err := rows.Scan(
			&i.ID,
			&i.Apartment.ID,
			&i.Address,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, i)
	}
	return images, nil
}

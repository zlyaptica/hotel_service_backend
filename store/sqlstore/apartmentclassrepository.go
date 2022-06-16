package sqlstore

import (
	"database/sql"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
	"github.com/zlyaptica/hotel_service_backend/store"
)

type ApartmentClassRepository struct {
	store *Store
}

func (r ApartmentClassRepository) FindAll() ([]model.ApartmentClass, error) {
	apartmentClasses := []model.ApartmentClass{}
	q := `SELECT id, class FROM apartment_classes`
	rows, err := r.store.db.Query(q)
	defer rows.Close()

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	for rows.Next() {
		ac := model.ApartmentClass{}
		err := rows.Scan(
			&ac.ID,
			&ac.Class,
		)
		if err != nil {
			return nil, err
		}

		apartmentClasses = append(apartmentClasses, ac)
	}
	return apartmentClasses, nil
}

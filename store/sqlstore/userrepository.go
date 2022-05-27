package sqlstore

import (
	"database/sql"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
	"github.com/zlyaptica/hotel_service_backend/store"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(u *model.User) error {
	q := `INSERT INTO guests (lname, fname, phone_number) VALUES ($1, $2, $3) RETURNING id`
	return r.store.db.QueryRow(
		q,
		u.LName,
		u.FName,
		u.PhoneNumber,
	).Scan(&u.ID)
}

func (r *UserRepository) Find(id int) (*model.User, error) {
	g := &model.User{}
	q := `SELECT id, lname, fname, phone_number FROM guests WHERE id = $1`
	if err := r.store.db.QueryRow(
		q,
		id,
	).Scan(
		&g.ID,
		&g.LName,
		&g.FName,
		&g.PhoneNumber,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}
	return g, nil
}

func (r *UserRepository) FindByPhone(phone string) (*model.User, error) {
	g := &model.User{}
	q := `SELECT id, lname, fname, phone_number FROM guests WHERE phone_number = $1`
	if err := r.store.db.QueryRow(
		q,
		phone,
	).Scan(
		&g.ID,
		&g.LName,
		&g.FName,
		&g.PhoneNumber,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}
	return g, nil
}

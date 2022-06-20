package sqlstore

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
	"github.com/zlyaptica/hotel_service_backend/store"
)

type UserRepository struct {
	store *Store
}

var (
	errUnknownPhoneNumber = errors.New("there is no user with this phone number")
)

func (r *UserRepository) Create(u *model.User) error {
	q := `INSERT INTO guests (lname, fname, phone_number) VALUES ($1, $2, $3) RETURNING id`
	return r.store.db.QueryRow(
		q,
		u.LName,
		u.FName,
		u.PhoneNumber,
	).Scan(&u.ID)
}

func (r *UserRepository) Delete(phoneNumber string) error {
	q := `DELETE FROM guests WHERE phone_number = $1`
	result, err := r.store.db.Exec(q, phoneNumber)
	if err != nil {
		return err
	}
	fmt.Println("result:", result, "err:", err)
	row, err := result.RowsAffected()
	if err != nil {
		return err
	}
	fmt.Println("row:", row, "err:", err)
	if row != 1 {
		return errUnknownPhoneNumber
	}
	return err
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

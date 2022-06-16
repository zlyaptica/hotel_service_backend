package sqlstore

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/zlyaptica/hotel_service_backend/store"
)

type Store struct {
	db                       *sql.DB
	addressRepository        *AddressRepository
	apartmentClassRepository *ApartmentClassRepository
	apartmentRepository      *ApartmentRepository
	userRepository           *UserRepository
	hotelRepository          *HotelRepository
	apartmentImageRepository *ApartmentImageRepository
	transactRepository       *TransactRepository
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) Address() store.AddressRepository {
	if s.addressRepository != nil {
		return s.addressRepository
	}

	s.addressRepository = &AddressRepository{
		store: s,
	}

	return s.addressRepository
}

func (s *Store) ApartmentClass() store.ApartmentClassRepository {
	if s.apartmentClassRepository != nil {
		return s.apartmentClassRepository
	}

	s.apartmentClassRepository = &ApartmentClassRepository{
		store: s,
	}

	return s.apartmentClassRepository
}

func (s *Store) Apartment() store.ApartmentRepository {
	if s.apartmentRepository != nil {
		return s.apartmentRepository
	}

	s.apartmentRepository = &ApartmentRepository{
		store: s,
	}

	return s.apartmentRepository
}

func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
	}

	return s.userRepository
}

func (s *Store) Hotel() store.HotelRepository {
	if s.hotelRepository != nil {
		return s.hotelRepository
	}

	s.hotelRepository = &HotelRepository{
		store: s,
	}

	return s.hotelRepository
}

func (s *Store) ApartmentImage() store.ApartmentImageRepository {
	if s.apartmentImageRepository != nil {
		return s.apartmentImageRepository
	}

	s.apartmentImageRepository = &ApartmentImageRepository{
		store: s,
	}

	return s.apartmentImageRepository
}

func (s *Store) Transact() store.TransactRepository {
	if s.transactRepository != nil {
		return s.transactRepository
	}

	s.transactRepository = &TransactRepository{
		store: s,
	}

	return s.transactRepository

}

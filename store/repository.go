package store

import (
	"encoding/json"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
)

type AddressRepository interface{}

type ApartmentClassRepository interface { // типа сделал
	FindAll() ([]model.ApartmentClass, error)
}

type ApartmentRepository interface {
	Create(bedCount, price, apartmentClassID, hotelID json.Number, name string) error
	GetPriceApartment(id int) (int, error)
	FillRoom(id int) error
	FindByHotelID(id int) ([]model.Apartment, error)
}

type UserRepository interface {
	Create(user *model.User) error
	Delete(phoneNumber string) error
	//Find(id int) (*model.User, error)
	//FindByPhone(string) (*model.User, error)
}

type HotelRepository interface {
	Create(hotel *model.Hotel) error
	Update(hotel *model.Hotel) error
	FindAll() ([]model.Hotel, error)
	Find(id int) (*model.Hotel, error)
}

type ApartmentImageRepository interface {
	GetImagesByHotelID(id int) ([]model.ApartmentImage, error)
}

type TransactRepository interface {
	Create(t *model.Transact) error
	CreateTransact(t *model.Transact) error
	FindTransactsByPhoneNumber(phoneNumber string) ([]model.Transact, error)
}

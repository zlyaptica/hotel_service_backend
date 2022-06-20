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
	Delete(id int) error
	FindAll() ([]model.Apartment, error)
	Find(id int) (*model.Apartment, error)
	GetPriceApartment(id int) (int, error)
	FillRoom(id int) error
	FindByHotelID(id int) ([]model.Apartment, error)
	FindByBedCount(count int) ([]model.Apartment, error)
	FindFree(bool) ([]model.Apartment, error)
	FindByClass(class string) ([]model.Apartment, error)
}

type UserRepository interface {
	Create(user *model.User) error
	Find(id int) (*model.User, error)
	FindByPhone(string) (*model.User, error)
	Delete(phoneNumber string) error
}

type HotelRepository interface {
	Create(hotel *model.Hotel) error
	Delete(id int) error
	Update(hotel *model.Hotel) error
	FindAll() ([]model.Hotel, error)
	Find(id int) (*model.Hotel, error)
	FindByCountry(country string) ([]model.Hotel, error)
	FindByCity(city string) ([]model.Hotel, error)
}

type ApartmentImageRepository interface {
	GetImagesByHotelID(id int) ([]model.ApartmentImage, error)
}

type TransactRepository interface {
	Create(t *model.Transact) error
	CreateTransact(t *model.Transact) error
	FindTransactsByPhoneNumber(phoneNumber string) ([]model.Transact, error)
}

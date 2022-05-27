package store

import "github.com/zlyaptica/hotel_service_backend/internal/app/model"

type AddressRepository interface{}

type ApartmentClassRepository interface { // типа сделал
	FindAll() ([]model.ApartmentClass, error)
}

// a *model.Apartment, apartmentClassID, hotelID int
type ApartmentRepository interface {
	Create(bedCount, price, apartmentClassID, hotelID int) error
	Delete(id int) error
	FindAll() ([]model.Apartment, error)
	Find(id int) (*model.Apartment, error)
	FindByHotel(hotel string) ([]model.Apartment, error)
	FindByBedCount(count int) ([]model.Apartment, error)
	FindFree(bool) ([]model.Apartment, error)
	FindByClass(class string) ([]model.Apartment, error)
}

type UserRepository interface { // типа сделал
	Create(user *model.User) error
	Find(id int) (*model.User, error)
	FindByPhone(string) (*model.User, error)
}

type HotelRepository interface { // сделал
	FindAll() ([]model.Hotel, error)
	Find(id int) (*model.Hotel, error)
	FindByCountry(country string) ([]model.Hotel, error)
	FindByCity(city string) ([]model.Hotel, error)
}

type ImageRepository interface {
	GetImages(id int) ([]model.Image, error)
}

type TransactRepository interface {
	Create(t *model.Transact) error
	FindTransactsByUserID(id int) ([]model.Transact, error)
}

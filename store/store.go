package store

type Store interface {
	Address() AddressRepository
	ApartmentClass() ApartmentClassRepository
	Apartment() ApartmentRepository
	User() UserRepository
	Hotel() HotelRepository
	ApartmentImage() ApartmentImageRepository
	Transact() TransactRepository
}

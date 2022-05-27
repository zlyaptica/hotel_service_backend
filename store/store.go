package store

type Store interface {
	Address() AddressRepository
	ApartmentClass() ApartmentClassRepository
	Apartment() ApartmentRepository
	User() UserRepository
	Hotel() HotelRepository
	Image() ImageRepository
	Transact() TransactRepository
}

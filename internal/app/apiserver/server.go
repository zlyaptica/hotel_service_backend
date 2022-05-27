package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/zlyaptica/hotel_service_backend/internal/app/model"
	"github.com/zlyaptica/hotel_service_backend/store"
	"net/http"
	"strconv"
	"time"
)

const (
	sessionName        = "hotelservice"
	ctxKeyUser  ctxKey = iota
	ctxKeyRequestID
)

type ctxKey int8

var (
	getApartmentClasses = "/apartmentclasses"

	createUsers   = "/users"
	createSession = "/sessions"
	whoami        = "/whoami"

	postTransact         = "/transacts"
	getTransactsByUserID = "/user/{id}/transacts"

	getHotels          = "/hotels"
	postHotels         = "/hotels"
	getHotel           = "/hotels/{id}"
	getHotelsByCountry = "/hotels/{country}"
	getHotelsByCity    = "/hotels/{city}"

	getApartments           = "/apartments"
	postApartments          = "/apartments"
	getApartment            = "/apartments/{id}"
	getApartmentsByHotel    = "/apartments/{hotel}"
	getApartmentsByBedCount = "/apartments/{bed_count}"
	getFreeApartments       = "/apartments/{is_free}"
	getApartmentsByClass    = "/apartments/{class}"

	getImages = "/apartments/{id}/images"

	errIncorrectNumber  = errors.New("incorrect number")
	errNotAuthenticated = errors.New("not authenticated")
)

type server struct {
	router       *mux.Router
	logger       *logrus.Logger
	store        store.Store
	sessionStore sessions.Store
}

func newServer(store store.Store, sessionStore sessions.Store) *server {
	s := &server{
		router:       mux.NewRouter(),
		logger:       logrus.New(),
		store:        store,
		sessionStore: sessionStore,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	s.router.HandleFunc(createUsers, s.handleUsersCreate()).Methods("POST")
	s.router.HandleFunc(createSession, s.handleSessionCreate()).Methods("POST")

	private := s.router.PathPrefix("/private").Subrouter()
	private.Use(s.authenticateUser)
	private.HandleFunc(whoami, s.handleWhoami()).Methods("GET")

	//
	private.HandleFunc(postTransact, s.handleTransactCreate()).Methods("POST")
	s.router.HandleFunc(getTransactsByUserID, s.handleTransactsGetByUserID()).Methods("POST")

	// ОТЕЛИ
	s.router.HandleFunc(getHotels, s.handleHotelsGet()).Methods("GET")
	s.router.HandleFunc(getHotel, s.handleHotelGet()).Methods("GET")
	s.router.HandleFunc(getHotelsByCountry, s.handleHotelsByCountryGet()).Methods("GET")
	s.router.HandleFunc(getHotelsByCity, s.handleHotelsByCityGet()).Methods("GET")

	// АПАРТАМЕНТЫ
	s.router.HandleFunc(postApartments, s.handleApartmentsCreate()).Methods("POST")
	s.router.HandleFunc(getApartments, s.handleApartmentsGet()).Methods("GET")
	s.router.HandleFunc(getApartment, s.handleApartmentGet()).Methods("GET")
	//s.router.HandleFunc(getApartmentsByHotel, s.handleApartmentsByHotelGet).Methods("GET")
	//s.router.HandleFunc(getApartmentsByBedCount, s.handleApartmentsByBedCountGet).Methods("GET")
	//s.router.HandleFunc(getFreeApartments, s.handleApartmentsByFreeGet).Methods("GET")
	//s.router.HandleFunc(getApartmentsByClass, s.handleApartmentsByClass).Methods("GET")

	// КЛАСС АПАРТАМЕНТА
	s.router.HandleFunc(getApartmentClasses, s.handleApartmentClassesGet()).Methods("GET")

	// КЛАСС КАРТИНОК
	s.router.HandleFunc(getImages, s.handleImagesGet()).Methods("GET")
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})

		logger.Infof("started %s %s", r.Method, r.RequestURI)
		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		logger.Infof(
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start),
		)
	})
}

func (s *server) handleUsersCreate() http.HandlerFunc {
	type request struct {
		LName       string `json:"l_name"`
		FName       string `json:"f_name"`
		PhoneNumber string `json:"phone_number"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		g := &model.User{
			LName:       req.LName,
			FName:       req.FName,
			PhoneNumber: req.PhoneNumber,
		}
		if err := s.store.User().Create(g); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusCreated, g)
	}
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		id, ok := session.Values["user_id"]
		if !ok {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}
		g, err := s.store.User().Find(id.(int))
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, g)))
	})
}

func (s *server) handleWhoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, r.Context().Value(ctxKeyUser).(*model.User))
	}
}

func (s *server) handleSessionCreate() http.HandlerFunc {
	type request struct {
		PhoneNumber string `json:"phone_number"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		g, err := s.store.User().FindByPhone(req.PhoneNumber)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errIncorrectNumber)
		}

		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		session.Values["user_id"] = g.ID
		if err := s.sessionStore.Save(r, w, session); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) handleTransactCreate() http.HandlerFunc {
	type request struct {
		ApartmentID   int       `json:"apartment_id"`
		DateArrival   time.Time `json:"date_arrival"`
		DateDeparture time.Time `json:"date_departure"`
		Price         int       `json:"price"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		user := r.Context().Value(ctxKeyUser).(*model.User)
		a := &model.Apartment{
			ID: req.ApartmentID,
		}
		t := &model.Transact{
			Apartment:     a,
			User:          user,
			DateArrival:   req.DateArrival,
			DateDeparture: req.DateDeparture,
			Price:         req.Price,
		}
		if err := s.store.Transact().Create(t); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
		}

		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) handleTransactsGetByUserID() http.HandlerFunc {
	type responce struct {
		Items []model.Transact `json:"items"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])

		transacts, err := s.store.Transact().FindTransactsByUserID(id)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}

		resp := &responce{
			Items: transacts,
		}

		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *server) handleApartmentClassesGet() http.HandlerFunc {
	type responce struct {
		Items []model.ApartmentClass `json:"items"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		apartmentClasses, err := s.store.ApartmentClass().FindAll()
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		resp := &responce{
			Items: apartmentClasses,
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *server) handleHotelsGet() http.HandlerFunc {
	type responce struct {
		Items []model.Hotel `json:"items"`
	}
	return func(w http.ResponseWriter, r *http.Request) {

		hotels, err := s.store.Hotel().FindAll()
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		resp := &responce{
			Items: hotels,
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *server) handleHotelGet() http.HandlerFunc {
	type responce struct {
		Item *model.Hotel `json:"item"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		hotel, err := s.store.Hotel().Find(id)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		resp := &responce{
			Item: hotel,
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *server) handleHotelsByCountryGet() http.HandlerFunc {
	type responce struct {
		Items []model.Hotel `json:"items"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		country := vars["country"]

		hotels, err := s.store.Hotel().FindByCountry(country)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		resp := &responce{
			Items: hotels,
		}

		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *server) handleHotelsByCityGet() http.HandlerFunc {
	type responce struct {
		Items []model.Hotel `json:"items"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		city := vars["city"]

		hotels, err := s.store.Hotel().FindByCity(city)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		resp := &responce{
			Items: hotels,
		}

		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *server) handleApartmentsCreate() http.HandlerFunc {
	type request struct {
		Name             string `json:"name"`
		BedCount         int    `json:"bed_count"`
		Price            int    `json:"price"`
		ApartmentClassID int    `json:"apartment_class_id"`
		HotelID          int    `json:"hotel_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		//vars := mux.Vars(r)
		//bedCount, err := strconv.Atoi(vars["bed_count"])
		//fmt.Println("bed count: ", bedCount)
		//if err != nil {
		//	s.error(w, r, http.StatusUnprocessableEntity, err)
		//	return
		//}
		//price, err := strconv.Atoi(vars["price"])
		//if err != nil {
		//	s.error(w, r, http.StatusUnprocessableEntity, err)
		//	return
		//}
		//apartmentClassID, err := strconv.Atoi(vars["apartment_class_id"])
		//if err != nil {
		//	s.error(w, r, http.StatusUnprocessableEntity, err)
		//	return
		//}
		//hotelID, err := strconv.Atoi(vars["hotel_id"])
		//if err != nil {
		//	s.error(w, r, http.StatusUnprocessableEntity, err)
		//	return
		//}
		//
		//ac := &model.ApartmentClass{
		//	ID: apartmentClassID,
		//}
		//h := &model.Hotel{
		//	ID: hotelID,
		//}
		//a := &model.Apartment{
		//	BedCount:       bedCount,
		//	Price:          price,
		//	Hotel:          h,
		//	ApartmentClass: ac,
		//}
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		if err := s.store.Apartment().Create(req.BedCount, req.Price, req.ApartmentClassID, req.HotelID); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		s.respond(w, r, http.StatusCreated, nil)
	}
}

func (s *server) handleApartmentsGet() http.HandlerFunc {
	type responce struct {
		Items []model.Apartment `json:"items"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		apartments, err := s.store.Apartment().FindAll()
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		resp := &responce{
			Items: apartments,
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *server) handleApartmentGet() http.HandlerFunc {
	type responce struct {
		Item *model.Apartment `json:"item"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		apartment, err := s.store.Apartment().Find(id)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		resp := &responce{
			Item: apartment,
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *server) handleImagesGet() http.HandlerFunc {
	type responce struct {
		Items []model.Image `json:"items"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		images, err := s.store.Image().GetImages(id)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		resp := &responce{
			Items: images,
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

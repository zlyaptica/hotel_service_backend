package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
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
	getTransactsByUserID = "/user/{phoneNumber}/transacts"

	getHotels = "/hotels"
	//getHotel  = "/hotels/{id}"
	//postHotel          = "/hotels"
	//getHotelsByCountry = "/hotels/{country}"
	//getHotelsByCity    = "/hotels/{city}"

	postApartments         = "/apartments"
	getApartmentsByHotelID = "/hotel/{id}/apartments"
	//getApartments           = "/apartments"
	//getApartment            = "/apartments/{id}"
	//getApartmentsByBedCount = "/apartments/{bed_count}"
	//getFreeApartments       = "/apartments/{is_free}"
	//getApartmentsByClass    = "/apartments/{class}"

	//getImages = "/apartments/{id}/images"

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
	s.router.Use(s.setCORS)
	s.router.HandleFunc(createUsers, s.handleUsersCreate()).Methods("POST", "OPTIONS")
	s.router.HandleFunc(createSession, s.handleSessionCreate()).Methods("POST")

	private := s.router.PathPrefix("/private").Subrouter()
	private.Use(s.authenticateUser)
	private.HandleFunc(whoami, s.handleWhoami()).Methods("GET")

	// ТРАНЗАКЦИИ
	s.router.HandleFunc(postTransact, s.handleTransactCreate()).Methods("POST", "OPTIONS")
	s.router.HandleFunc(getTransactsByUserID, s.handleTransactsGetByUserID()).Methods("GET")

	// ОТЕЛИ
	s.router.HandleFunc(getHotels, s.handleHotelsGet()).Methods("GET")
	//s.router.HandleFunc(getHotel, s.handleHotelGet()).Methods("GET")
	//s.router.HandleFunc(getHotelsByCountry, s.handleHotelsByCountryGet()).Methods("GET")
	//s.router.HandleFunc(getHotelsByCity, s.handleHotelsByCityGet()).Methods("GET")
	//s.router.HandleFunc(postHotel, s.handleHotelCreate()).Methods("POST")

	// АПАРТАМЕНТЫ
	s.router.HandleFunc(postApartments, s.handleApartmentsCreate()).Methods("POST", "OPTIONS")
	s.router.HandleFunc(getApartmentsByHotelID, s.handleApartmentsByHotelIDGet()).Methods("GET")
	//s.router.HandleFunc(getApartments, s.handleApartmentsGet()).Methods("GET")
	//s.router.HandleFunc(getApartment, s.handleApartmentGet()).Methods("GET")
	//s.router.HandleFunc(getApartmentsByBedCount, s.handleApartmentsByBedCountGet).Methods("GET")
	//s.router.HandleFunc(getFreeApartments, s.handleApartmentsByFreeGet).Methods("GET")
	//s.router.HandleFunc(getApartmentsByClass, s.handleApartmentsByClass).Methods("GET")

	// КЛАСС АПАРТАМЕНТА
	s.router.HandleFunc(getApartmentClasses, s.handleApartmentClassesGet()).Methods("GET")

	// КЛАСС КАРТИНОК
	//s.router.HandleFunc(getImages, s.handleImagesGet()).Methods("GET")
}

func (s *server) setCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
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
		u := &model.User{
			LName:       req.LName,
			FName:       req.FName,
			PhoneNumber: req.PhoneNumber,
		}
		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, r, http.StatusCreated, u)
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
		PhoneNumber   string `json:"phone_number"`
		ApartmentID   int    `json:"apartment_id"`
		DateArrival   string `json:"date_arrival"`
		DateDeparture string `json:"date_departure"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		fmt.Println("req > ", req)
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			fmt.Println("err > ", err)
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		u := &model.User{
			PhoneNumber: req.PhoneNumber,
		}
		a := &model.Apartment{
			ID: req.ApartmentID,
		}

		dateArrival, err := time.Parse("2006-01-02", req.DateArrival)
		if err != nil {
			fmt.Println("invalid date arrival:", err)
			return
		}
		dateDeparture, err := time.Parse("2006-01-02", req.DateDeparture)
		if err != nil {
			fmt.Println("invalid date departure", err)
			return
		}

		dur := dateDeparture.Sub(dateArrival).Hours() / 24
		apartmentPrice, err := s.store.Apartment().GetPriceApartment(req.ApartmentID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		finishPrice := dur * float64(apartmentPrice)
		t := &model.Transact{
			Apartment:     a,
			User:          u,
			DateArrival:   dateArrival,
			DateDeparture: dateDeparture,
			Price:         int(finishPrice),
		}
		if err := s.store.Transact().CreateTransact(t); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		err = s.store.Apartment().FillRoom(req.ApartmentID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
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
		phoneNumber := vars["phoneNumber"]
		fmt.Println("phoneNumber > ", phoneNumber)
		transacts, err := s.store.Transact().FindTransactsByPhoneNumber(phoneNumber)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
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
	type response struct {
		Hotels []model.Hotel `json:"hotels"`
	}
	return func(w http.ResponseWriter, r *http.Request) {

		hotels, err := s.store.Hotel().FindAll()
		if err != nil {
			fmt.Println("err:", err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		resp := &response{
			Hotels: hotels,
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

//func (s *server) handleHotelGet() http.HandlerFunc {
//	type responce struct {
//		Hotel  *model.Hotel           `json:"hotel"`
//		Images []model.ApartmentImage `json:"images"`
//	}
//	return func(w http.ResponseWriter, r *http.Request) {
//		vars := mux.Vars(r)
//		id, err := strconv.Atoi(vars["id"])
//		hotel, err := s.store.Hotel().Find(id)
//		if err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//		images, err := s.store.ApartmentImage().GetImages()
//		if err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//		resp := &responce{
//			Hotel:  hotel,
//			Images: images,
//		}
//		s.respond(w, r, http.StatusOK, resp)
//	}
//}

//func (s *server) handleHotelCreate() http.HandlerFunc {
//	type request struct {
//		Name        string `json:"name"`
//		StarsCount  int    `json:"stars_count"`
//		Description string `json:"description"`
//		Country     string `json:"country"`
//		City        string `json:"city"`
//		Street      string `json:"street"`
//		House       string `json:"house"`
//	}
//	return func(w http.ResponseWriter, r *http.Request) {
//		vars := mux.Vars(r)
//		id, err := strconv.Atoi(vars["id"])
//		hotel, err := s.store.Hotel().Find(id)
//		if err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//		resp := &responce{
//			Item: hotel,
//		}
//		s.respond(w, r, http.StatusOK, resp)
//	}
//}

//func (s *server) handleHotelsByCountryGet() http.HandlerFunc {
//	type responce struct {
//		Items []model.Hotel `json:"items"`
//	}
//	return func(w http.ResponseWriter, r *http.Request) {
//		vars := mux.Vars(r)
//		country := vars["country"]
//
//		hotels, err := s.store.Hotel().FindByCountry(country)
//		if err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//
//		resp := &responce{
//			Items: hotels,
//		}
//
//		s.respond(w, r, http.StatusOK, resp)
//	}
//}
//
//func (s *server) handleHotelsByCityGet() http.HandlerFunc {
//	type responce struct {
//		Items []model.Hotel `json:"items"`
//	}
//	return func(w http.ResponseWriter, r *http.Request) {
//		vars := mux.Vars(r)
//		city := vars["city"]
//
//		hotels, err := s.store.Hotel().FindByCity(city)
//		if err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//
//		resp := &responce{
//			Items: hotels,
//		}
//
//		s.respond(w, r, http.StatusOK, resp)
//	}
//}

func (s *server) handleApartmentsCreate() http.HandlerFunc {
	type request struct {
		Name             string      `json:"name"`
		BedCount         json.Number `json:"bed_count"`
		Price            json.Number `json:"price"`
		ApartmentClassID json.Number `json:"apartment_class_id"`
		HotelID          json.Number `json:"hotel_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			fmt.Println("err > ", err)
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		if err := s.store.Apartment().Create(req.BedCount, req.Price, req.ApartmentClassID, req.HotelID, req.Name); err != nil {
			fmt.Println("err > ", err)
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		s.respond(w, r, http.StatusCreated, nil)
	}
}

//func (s *server) handleApartmentsGet() http.HandlerFunc {
//	type responce struct {
//		Items []model.Apartment `json:"items"`
//	}
//	return func(w http.ResponseWriter, r *http.Request) {
//		apartments, err := s.store.Apartment().FindAll()
//		if err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//		resp := &responce{
//			Items: apartments,
//		}
//		s.respond(w, r, http.StatusOK, resp)
//	}
//}

//func (s *server) handleApartmentGet() http.HandlerFunc {
//	type responce struct {
//		Item *model.Apartment `json:"item"`
//	}
//	return func(w http.ResponseWriter, r *http.Request) {
//		vars := mux.Vars(r)
//		id, err := strconv.Atoi(vars["id"])
//		apartment, err := s.store.Apartment().Find(id)
//		if err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//		resp := &responce{
//			Item: apartment,
//		}
//		s.respond(w, r, http.StatusOK, resp)
//	}
//}

func (s *server) handleApartmentsByHotelIDGet() http.HandlerFunc {
	type response struct {
		Apartments       []model.Apartment      `json:"apartments"`
		ApartmentsImages []model.ApartmentImage `json:"apartments_images"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])

		apartments, err := s.store.Apartment().FindByHotelID(id)
		if err != nil {
			fmt.Println("err > ", err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		apartmentsImages, err := s.store.ApartmentImage().GetImagesByHotelID(id)
		if err != nil {
			fmt.Println("err > ", err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		resp := &response{
			Apartments:       apartments,
			ApartmentsImages: apartmentsImages,
		}
		s.respond(w, r, http.StatusOK, resp)
	}
}

//func (s *server) handleImagesGet() http.HandlerFunc {
//	type responce struct {
//		Items []model.ApartmentImage `json:"items"`
//	}
//	return func(w http.ResponseWriter, r *http.Request) {
//		//vars := mux.Vars(r)
//		//id, err := strconv.Atoi(vars["id"])
//		images, err := s.store.ApartmentImage().GetImagesByApartmentID(id)
//		if err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//		resp := &responce{
//			Items: images,
//		}
//		s.respond(w, r, http.StatusOK, resp)
//	}
//}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

package apiserver

import (
	"context"
	"encoding/json"
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

	createUsers = "/users"
	deleteUsers = "/users/{phone_number}"
	//createSession = "/sessions"
	//whoami        = "/whoami"

	postTransact         = "/transacts"
	getTransactsByUserID = "/user/{phoneNumber}/transacts"

	getHotels   = "/hotels"
	getHotel    = "/hotels/{id}"
	createHotel = "/hotels"
	updateHotel = "/hotels/{id}"
	//deleteHotel = "/hotels/{id}"
	//getHotelsByCountry = "/hotels/{country}"
	//getHotelsByCity    = "/hotels/{city}"

	postApartments         = "/apartments"
	getApartmentsByHotelID = "/hotel/{id}/apartments"

	//errNotAuthenticated = errors.New("not authenticated")
	//errIncorrectNumber  = errors.New("incorrect number")
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
	s.router.HandleFunc(deleteUsers, s.handleUsersDelete()).Methods("DELETE", "OPTIONS")
	//s.router.HandleFunc(createSession, s.handleSessionCreate()).Methods("POST", "OPTIONS")

	//private := s.router.PathPrefix("/private").Subrouter()
	//private.Use(s.authenticateUser)
	//private.HandleFunc(whoami, s.handleWhoami()).Methods("GET")

	// ТРАНЗАКЦИИ
	s.router.HandleFunc(postTransact, s.handleTransactCreate()).Methods("POST", "OPTIONS")
	s.router.HandleFunc(getTransactsByUserID, s.handleTransactsGetByUserID()).Methods("GET")

	// ОТЕЛИ
	s.router.HandleFunc(getHotels, s.handleHotelsGet()).Methods("GET") // хэндлер на путь localhost:8080/hotels
	// с методом GET
	s.router.HandleFunc(getHotel, s.handleHotelGet()).Methods("GET")
	s.router.HandleFunc(createHotel, s.handleHotelCreate()).Methods("POST", "OPTIONS")
	s.router.HandleFunc(updateHotel, s.handleHotelUpdate()).Methods("PUT", "OPTIONS")

	// АПАРТАМЕНТЫ
	s.router.HandleFunc(postApartments, s.handleApartmentsCreate()).Methods("POST", "OPTIONS")
	s.router.HandleFunc(getApartmentsByHotelID, s.handleApartmentsByHotelIDGet()).Methods("GET")

	// КЛАСС АПАРТАМЕНТА
	s.router.HandleFunc(getApartmentClasses, s.handleApartmentClassesGet()).Methods("GET")
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

//func (s *server) authenticateUser(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		session, err := s.sessionStore.Get(r, sessionName)
//		if err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//
//		id, ok := session.Values["user_id"]
//		if !ok {
//			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
//			return
//		}
//		g, err := s.store.User().Find(id.(int))
//		if err != nil {
//			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
//			return
//		}
//		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, g)))
//	})
//}

//func (s *server) handleWhoami() http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		s.respond(w, r, http.StatusOK, r.Context().Value(ctxKeyUser).(*model.User))
//	}
//}

//func (s *server) handleSessionCreate() http.HandlerFunc {
//	type request struct {
//		PhoneNumber string `json:"phone_number"`
//	}
//
//	return func(w http.ResponseWriter, r *http.Request) {
//		req := &request{}
//		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
//			s.error(w, r, http.StatusBadRequest, err)
//			return
//		}
//		g, err := s.store.User().FindByPhone(req.PhoneNumber)
//		if err != nil {
//			s.error(w, r, http.StatusUnauthorized, errIncorrectNumber)
//		}
//
//		session, err := s.sessionStore.Get(r, sessionName)
//		if err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//
//		session.Values["user_id"] = g.ID
//		if err := s.sessionStore.Save(r, w, session); err != nil {
//			s.error(w, r, http.StatusInternalServerError, err)
//			return
//		}
//
//		s.respond(w, r, http.StatusOK, nil)
//	}
//}

func (s *server) handleUsersDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		phoneNumber := vars["phone_number"]
		if err := s.store.User().Delete(phoneNumber); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
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
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
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
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		dateDeparture, err := time.Parse("2006-01-02", req.DateDeparture)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
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
	type response struct {
		Items []model.Transact `json:"items"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		phoneNumber := vars["phoneNumber"]
		transacts, err := s.store.Transact().FindTransactsByPhoneNumber(phoneNumber)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		resp := &response{
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
	type response struct { // структура с массивом отелей для отправки на сайт
		Hotels []model.Hotel `json:"hotels"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		hotels, err := s.store.Hotel().FindAll() // в отели получаем массив отелей,
		// в ошибку - ошибку
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return // если есть ошибка, то логируем ее с 500 ошибкой и выходим с функции
		}
		resp := &response{
			Hotels: hotels, // если все ок, то добавляем отели в структуру, которая отправится на сайт
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

func (s *server) handleHotelCreate() http.HandlerFunc {
	type request struct {
		Name               string `json:"name"`
		StarsCount         int    `json:"stars_count"`
		Description        string `json:"description"`
		Country            string `json:"country"`
		City               string `json:"city"`
		Street             string `json:"street"`
		House              string `json:"house"`
		HeaderImageAddress string `json:"header_image_address"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		a := &model.Address{
			Country: req.Country,
			City:    req.City,
			Street:  req.Street,
			House:   req.House,
		}
		h := &model.Hotel{
			Name:               req.Name,
			Address:            a,
			StarsCount:         req.StarsCount,
			Description:        req.Description,
			HeaderImageAddress: req.HeaderImageAddress,
		}
		if err := s.store.Hotel().Create(h); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		s.respond(w, r, http.StatusCreated, h)
	}
}

func (s *server) handleHotelUpdate() http.HandlerFunc {
	type request struct {
		Name               string `json:"name"`
		StarsCount         int    `json:"stars_count"`
		Description        string `json:"description"`
		Country            string `json:"country"`
		City               string `json:"city"`
		Street             string `json:"street"`
		House              string `json:"house"`
		HeaderImageAddress string `json:"header_image_address"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		fmt.Println("id in server.go = ", id)
		if err != nil {

			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		_, err = s.store.Hotel().Find(id)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		a := &model.Address{
			Country: req.Country,
			City:    req.City,
			Street:  req.Street,
			House:   req.House,
		}
		h := &model.Hotel{
			ID:                 id,
			Name:               req.Name,
			Address:            a,
			StarsCount:         req.StarsCount,
			Description:        req.Description,
			HeaderImageAddress: req.HeaderImageAddress,
		}
		fmt.Println("че по чем")
		err = s.store.Hotel().Update(h)
		if err != nil {
			fmt.Println("err = ", err)
		}
		s.respond(w, r, http.StatusOK, nil)
	}
}

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
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		s.respond(w, r, http.StatusCreated, nil)
	}
}

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

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

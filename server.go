package main

import (
	e "errors"
	"fmt"
	"net/http"
	"strings"

	fft "./api"
	auth "github.com/auth0-community/go-auth0"
	"github.com/go-chi/chi"
	cors "github.com/go-chi/cors"
	"github.com/go-chi/render"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

var validator *auth.JWTValidator

func main() {
	validator = getValidator()

	r := chi.NewRouter()

	r.Use(getCors().Handler)

	r.Route("/api/public", func(r chi.Router) {
		r.Get("/rooms", getAllRooms)
	})

	r.Route("/api/private", func(r chi.Router) {
		r.Use(validateJwtToken(validator))
		r.Post("/freetimes", checkFreeTimes)
	})

	r.Route("/api/limitedprivate", func(r chi.Router) {
		r.Use(validateJwtTokenAndScope(validator))
		r.Get("/history", getHistory)
	})

	http.ListenAndServe(":3000", r)
}

//==============================
// endpoints (start)
//==============================

// GET /api/public/rooms
func getAllRooms(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, NewAllRoomsResponse(fft.GetAllRooms()))
}

// GET /api/limitedprivate/history
func getHistory(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, NewHistoryResponse(fft.GetHistory()))
}

// POST /api/private/freetimes
func checkFreeTimes(w http.ResponseWriter, r *http.Request) {
	data := &FreeTimesRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	roomTimes, fftErr := fft.Find(data.Weekday, data.StartTime, data.EndTime, data.Rooms)

	if fftErr != nil {
		render.Render(w, r, ErrFFT(fftErr))
		return
	}

	render.Render(w, r, NewFreeTimesResponse(roomTimes))
}

//==============================
// endpoints (end)
//==============================

//==============================
// middleware configs (start)
//==============================

func getValidator() *auth.JWTValidator {
	client := auth.NewJWKClient(auth.JWKClientOptions{URI: "https://thailekha.auth0.com/.well-known/jwks.json"})
	audience := "http://localhost:3000"
	configuration := auth.NewConfiguration(client, []string{audience}, "https://thailekha.auth0.com/", jose.RS256)
	return auth.NewValidator(configuration)
}

func getCors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
}

/**
 * Input: JWTValidator, Output a func that receives a http.Handler and returns a http.Handler
 */
func validateJwtToken(validator *auth.JWTValidator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			_, err := validator.ValidateRequest(r)

			if err != nil {
				render.Render(w, r, ErrUnauthorizedRequest(err))
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func validateJwtTokenAndScope(validator *auth.JWTValidator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			token, err := validator.ValidateRequest(r)

			if err != nil {
				render.Render(w, r, ErrUnauthorizedRequest(err))
				return
			}

			if !hasSufficientScope(r, validator, token) {
				render.Render(w, r, ErrInsufficientScopeRequest(e.New("You do not have the read:history scope.")))
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// https://auth0.com/docs/quickstart/backend/golang/01-authorization
func hasSufficientScope(r *http.Request, validator *auth.JWTValidator, token *jwt.JSONWebToken) bool {
	claims := map[string]interface{}{}
	err := validator.Claims(r, token, &claims)

	if err != nil {
		fmt.Println(err)
	}

	return err == nil && strings.Contains(claims["scope"].(string), "read:history")
}

//==============================
// middleware configs (end)
//==============================

//============================
// Request related (start)
//============================

type FreeTimesRequest struct {
	Weekday   string
	StartTime string
	EndTime   string
	Rooms     []string
}

func (f *FreeTimesRequest) Bind(r *http.Request) error {
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrUnauthorizedRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 403,
		StatusText:     "Unauthorized request.",
		ErrorText:      err.Error(),
	}
}

func ErrInsufficientScopeRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 401,
		StatusText:     "Insufficient scope",
		ErrorText:      err.Error(),
	}
}

//============================
// Request related (end)
//============================

//============================
// Response related (start)
//============================

type FreeTimesResponse struct {
	Rooms []fft.RoomTimes `json:"rooms"`
}

type AllRoomsResponse struct {
	Rooms []string `json:"rooms"`
}

type HistoryResponse struct {
	History string `json:"history"`
}

func NewFreeTimesResponse(roomTimes []fft.RoomTimes) *FreeTimesResponse {
	return &FreeTimesResponse{roomTimes}
}

func NewAllRoomsResponse(rooms []string) *AllRoomsResponse {
	return &AllRoomsResponse{rooms}
}

func NewHistoryResponse(history string) *HistoryResponse {
	return &HistoryResponse{history}
}

func (ft *FreeTimesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (al *AllRoomsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (h *HistoryResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrFFT(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error finding free times",
		ErrorText:      err.Error(),
	}
}

//============================
// Response related (end)
//============================

//============================
// Helpers (start)
//============================

//============================
// Helpers (end)
//============================

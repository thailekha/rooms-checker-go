package main

import (
	"fmt"
	"net/http"

	fft "./api"
	auth "github.com/auth0-community/go-auth0"
	"github.com/go-chi/chi"
	cors "github.com/go-chi/cors"
	"github.com/go-chi/docgen"
	"github.com/go-chi/render"
	"gopkg.in/square/go-jose.v2"
)

// weekday, start time, end time
// weekday, start time, end time, list of rooms

var validator *auth.JWTValidator

func main() {
	validator = getValidator()

	r := chi.NewRouter()

	r.Use(getCors().Handler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Route("/api/public", func(r chi.Router) {
		r.Get("/rooms", getAllRooms)
	})

	r.Route("/api/private", func(r chi.Router) {
		r.Use(JwtMiddleware(validator))
		r.Post("/freetimes", checkFreeTimes)
	})

	fmt.Println(docgen.JSONRoutesDoc(r))

	fmt.Println("Listening at 3000")
	http.ListenAndServe(":3000", r)
}

//==============================
// endpoints (start)
//==============================

func checkFreeTimes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Token validated")

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

	fmt.Println("Done fetching, responding to client ...")
	render.Render(w, r, NewFreeTimesResponse(roomTimes))
}

func getAllRooms(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, NewAllRoomsResponse(fft.GetAllRooms()))
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
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
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
func JwtMiddleware(validator *auth.JWTValidator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r)
			token, err := validator.ValidateRequest(r)

			if err != nil {
				fmt.Println(err)
				fmt.Println("Token is not valid:", token)
				render.Render(w, r, ErrUnauthorizedRequest(err))
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
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

type AllRoomsRequest struct {
}

func (f *FreeTimesRequest) Bind(r *http.Request) error {
	return nil
}

func (f *AllRoomsRequest) Bind(r *http.Request) error {
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

//============================
// Request related (end)
//============================

//============================
// Response related (start)
//============================

type FreeTimesResponse struct {
	Rooms []fft.RoomTimes `json:"rooms"` //when this is encoded in json it will have key freeTimes instead of FreeTimes
}

type AllRoomsResponse struct {
	Rooms []string `json:"rooms"` //when this is encoded in json it will have key freeTimes instead of FreeTimes
}

func NewFreeTimesResponse(roomTimes []fft.RoomTimes) *FreeTimesResponse {
	return &FreeTimesResponse{roomTimes}
}

func NewAllRoomsResponse(rooms []string) *AllRoomsResponse {
	return &AllRoomsResponse{rooms}
}

func (ft *FreeTimesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	//preconfigure before encoding to json
	return nil
}

func (al *AllRoomsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	//preconfigure before encoding to json
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

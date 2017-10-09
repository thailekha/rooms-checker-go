package main

import (
	fft "./api"
	"fmt"
	"github.com/go-chi/chi"
	cors "github.com/go-chi/cors"
	"github.com/go-chi/docgen"
	"github.com/go-chi/render"
	"net/http"
)

// weekday, start time, end time
// weekday, start time, end time, list of rooms

func print(s string) {
	fmt.Println(s)
}

func main() {
	r := chi.NewRouter()

	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/freetimes", checkFreeTimes)
	})

	print(docgen.JSONRoutesDoc(r))

	print("Listening at 3000")
	http.ListenAndServe(":3000", r)
}

func checkFreeTimes(w http.ResponseWriter, r *http.Request) {
	print("PING")

	data := &FreeTimesRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	roomTimes, fftErr := fft.Find(data.Weekday, data.StartTime, data.EndTime, true)

	if fftErr != nil {
		render.Render(w, r, ErrFFT(fftErr))
		return
	}

	render.Render(w, r, NewFreeTimesResponse(roomTimes))
}

//============================
// Request related (start)
//============================

type FreeTimesRequest struct {
	Weekday   string
	StartTime string
	EndTime   string
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

//============================
// Request related (end)
//============================

//============================
// Response related (start)
//============================

type FreeTimesResponse struct {
	freeTimes []fft.RoomTimes
}

func NewFreeTimesResponse(roomTimes []fft.RoomTimes) *FreeTimesResponse {
	return &FreeTimesResponse{roomTimes}
}

func (ft *FreeTimesResponse) Render(w http.ResponseWriter, r *http.Request) error {
	for _, rts := range ft.freeTimes {
		fft.PrintRTS(rts)
	}
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

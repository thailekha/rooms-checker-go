package main

import (
	fft "./findfreetimes"
	"fmt"
	"github.com/go-chi/chi"
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
	data := &FreeTimesRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	freeTimes, fftErr := fft.Find(data.Weekday, data.StartTime, data.EndTime, true)

	if fftErr != nil {
		render.Render(w, r, ErrFFT(fftErr))
		return
	}

	render.Render(w, r, NewFreeTimesResponse(freeTimes))
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
	FreeTimes string
}

func NewFreeTimesResponse(freeTimes string) *FreeTimesResponse {
	return &FreeTimesResponse{FreeTimes: freeTimes}
}

func (ft *FreeTimesResponse) Render(w http.ResponseWriter, r *http.Request) error {
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

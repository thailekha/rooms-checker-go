package main

import (
	fft "./findfreetimes"
	"fmt"
	"github.com/go-chi/chi"
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

	r.Post("/api/freetimes", checkFreeTimes)

	print("Listening at 3000")
	http.ListenAndServe(":3000", r)
}

func checkFreeTimes(w http.ResponseWriter, r *http.Request) {
	data := &FreeTimesRequest{}
	if err := render.Bind(r, data); err != nil {
		w.Write([]byte("error"))
		return
	}

	print(data.Weekday)
	print(data.StartTime)
	print(data.EndTime)

	freeTimes, fftErr := fft.Find(data.Weekday, data.StartTime, data.EndTime, true)

	if fftErr != nil {
		w.Write([]byte(fftErr.Error()))
	}

	print(freeTimes)

	w.Write([]byte("cool"))
}

type FreeTimesRequest struct {
	Weekday   string
	StartTime string
	EndTime   string
}

func (f *FreeTimesRequest) Bind(r *http.Request) error {
	return nil
}

package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tamalsaha/learn-chi/chim"
	"k8s.io/klog/v2/klogr"
	"net/http"
	"time"
)

func main() {
	// Routes
	r := chi.NewRouter()
	r.Use(chim.RequestID)
	r.Use(chim.NewLogr(klogr.New().WithName("chi"), nil))
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Get("/wait", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		chim.LogEntrySetField(r, "wait", true)
		w.Write([]byte("hi"))
	})
	r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("oops")
	})
	http.ListenAndServe(":3333", r)
}

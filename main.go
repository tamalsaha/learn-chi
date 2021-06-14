package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tamalsaha/learn-chi/binding"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(binding.Injector)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})
	r.Get("/inject", binding.H(hello))

	log.Println("running server on :3333")
	http.ListenAndServe(":3333", r)
}

func hello(r *http.Request) string {
	return "hello " + r.URL.Query().Get("name")
}

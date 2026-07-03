package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	r := mux.NewRouter()

	r.PathPrefix("/docs/").Handler(
		http.StripPrefix("/docs/", http.FileServer(http.Dir("./combined"))),
	)

	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/docs/combined.yaml"),
	))

	fmt.Println("test")
	http.ListenAndServe(":8090", r)
}

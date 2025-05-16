package main

import (
	order "clients/api/handler"
	"clients/api/middleware"
	repository "clients/reposiroty"
	"log"
	"net/http"
)

func main() {
	db, err := repository.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()

	// repository
	orderRepo := repository.NewOrderRepository(db)

	// Handlers
	order.NewAuthHandler(router, *orderRepo)

	// Middlewares
	stack := middleware.Chain(
		middleware.CORS,
	)

	// Server
	server := http.Server{
		Addr:    ":8088",
		Handler: stack(router),
	}

	log.Printf("Starting HTTP server on port %s\n", ":8088")
	server.ListenAndServe()

}

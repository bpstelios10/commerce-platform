package main

import (
	"fmt"
	"net/http"

	httpx "commerce-platform/services/orders/internal/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Println("Commerce Platform - ORDERS")

	healthHandler := httpx.NewHealthHandler()

	r := chi.NewRouter()

	healthHandler.RegisterRoutes(r)

	http.ListenAndServe(":8080", r)
}

package order

import (
	"clients/api/helpers"
	repository "clients/reposiroty"
	"context"
	"strconv"

	"net/http"
)

type authHandler struct {
	repo repository.OrderRepository
}

func NewAuthHandler(router *http.ServeMux, repo repository.OrderRepository) {
	handler := &authHandler{
		repo: repo,
	}

	router.HandleFunc("GET /api/get-order/{id}", handler.GetOrder())

}

func (h *authHandler) GetOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		id := r.PathValue("id")
		idd, _ := strconv.Atoi(id)

		order, _ := h.repo.GetOrderRest(ctx, idd)

		helpers.Json(w, order, http.StatusOK)
	}
}

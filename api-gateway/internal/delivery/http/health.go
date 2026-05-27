package httpdelivery

import (
	"net/http"

	"food_delivery_platform/shared/utils"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func LiveHandler(w http.ResponseWriter, r *http.Request) {
	_ = utils.WriteJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Service: "api-gateway",
	})
}

func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	_ = utils.WriteJSON(w, http.StatusOK, healthResponse{
		Status:  "ready",
		Service: "api-gateway",
	})
}

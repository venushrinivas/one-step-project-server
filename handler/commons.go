package handler

import (
	"main/data"
	"net/http"
)

const OneStepDeviceApiUrl = "https://track.onestepgps.com/v3/api/public/device?latest_point=true&api-key=%s"

type ApiResponse struct {
	Devices []Device `json:"result_list"`
}

type Response struct {
	Message string `json:"message"`
}

type Handler struct {
	Preferences data.Preferences
	httpClient  *http.Client
}

func NewHandler(p data.Preferences, client *http.Client) *Handler {
	return &Handler{Preferences: p, httpClient: client}
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

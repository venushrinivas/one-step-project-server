package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/data"
	"net/http"
	"os"
	"path/filepath"
)

func (h *Handler) PreferencesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dataField := r.Form.Get("data")
		err = json.Unmarshal([]byte(dataField), &h.Preferences)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = h.Preferences.Save()
		if err != nil {
			log.Println("Error while persisting preferences")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response := Response{Message: "Request processed successfully"}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)
	} else if r.Method == http.MethodGet {
		apiUrl := fmt.Sprintf(OneStepDeviceApiUrl, os.Getenv("API_KEY"))
		res, err := h.httpClient.Get(apiUrl)
		defer res.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var resultList ApiResponse

		err = json.NewDecoder(res.Body).Decode(&resultList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var devicePreferences = make([]data.DevicePreferences, 0)
		for _, device := range resultList.Devices {
			matched := data.DevicePreferences{
				DeviceID:    device.DeviceID,
				DisplayName: device.DisplayName,
				Hidden:      false,
				Image:       "/images/default.png",
			}
			for _, devicePreference := range h.Preferences.GetDevicePreferences() {
				if devicePreference.DeviceID == device.DeviceID {
					matched.Hidden = devicePreference.Hidden
					matched.Image = devicePreference.Image
					break
				}
			}
			devicePreferences = append(devicePreferences, matched)
		}
		err = h.Preferences.SetDevicePreferences(devicePreferences)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h.Preferences)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == http.MethodPost {
		file, header, err := r.FormFile("file")
		if header != nil {
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		queryParams := r.URL.Query()
		deviceId := queryParams.Get("device_id")
		imageFilePath := "images/" + deviceId + filepath.Ext(header.Filename)
		serverFile, err := os.Create(imageFilePath)
		defer serverFile.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(serverFile, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for idx, _ := range h.Preferences.GetDevicePreferences() {
			if h.Preferences.GetDevicePreferences()[idx].DeviceID == deviceId {
				h.Preferences.GetDevicePreferences()[idx].Image = "/" + imageFilePath
				break
			}
		}
		h.Preferences.Save()
		response := Response{Message: "/" + imageFilePath}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

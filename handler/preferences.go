package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"main/data"
	"net/http"
	"os"
	"path/filepath"
)

// PreferencesHandler is the handler function for the preferences api call, handles both get and post request.
// Accepts a request and response object
func (h *Handler) PreferencesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Extracting form data present in data field
		dataField := r.Form.Get("data")
		// Deserializing the data into a preferences object
		err = json.Unmarshal([]byte(dataField), &h.Preferences)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Saving the preferences to the storage
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
		// Formatting the onestep api url by appending the API_KEY
		apiUrl := fmt.Sprintf(OneStepDeviceApiUrl, os.Getenv("API_KEY"))
		res, err := h.httpClient.Get(apiUrl)
		defer res.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var resultList ApiResponse
		// Deserializing the api response into a resultList object
		err = json.NewDecoder(res.Body).Decode(&resultList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Creating a device preferences array which holds individual device preferences
		var devicePreferences = make([]data.DevicePreferences, 0)
		for _, device := range resultList.Devices {
			matched := data.DevicePreferences{
				DeviceID:    device.DeviceID,
				DisplayName: device.DisplayName,
				Hidden:      false,
				Image:       DefaultImagePath,
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
		// Updating the device preferences to include the new devices which could have been added
		err = h.Preferences.SetDevicePreferences(devicePreferences)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h.Preferences)
	} else {
		// Handling error for other method types other than get and post
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// Upload is the method which handles image uploads. Accepts a request and response object
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	enableCors(w)
	if r.Method == http.MethodPost {
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		// extracting device_id from query params
		queryParams := r.URL.Query()
		deviceId := queryParams.Get("device_id")
		// Create a directory if it doesn't exist
		err = h.FileSystem.MkdirAll("images", os.ModePerm)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Constructing image file path
		imageFilePath := "images/" + deviceId + filepath.Ext(header.Filename)
		// Creating a file in the server to hold the image
		serverFile, err := h.FileSystem.Create(imageFilePath)
		defer serverFile.Close()
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Copying the uploaded image to the newly created file
		_, err = h.FileSystem.Copy(serverFile, file)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// setting the updated image in the device preferences
		for idx, _ := range h.Preferences.GetDevicePreferences() {
			if h.Preferences.GetDevicePreferences()[idx].DeviceID == deviceId {
				h.Preferences.GetDevicePreferences()[idx].Image = "/" + imageFilePath
				break
			}
		}
		// Saving the preferences to storage
		h.Preferences.Save()
		response := Response{Message: "/" + imageFilePath}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)
	} else {
		// Handling error for other http methods
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

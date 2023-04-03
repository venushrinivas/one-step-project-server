package main

import (
	"log"
	"main/data"
	"main/handler"
	"net/http"
	"os"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	port := os.Getenv("PORT")
	if apiKey == "" {
		log.Fatal("API_KEY environment is not set")
	}

	if port == "" {
		port = "8081"
	}
	_, err := os.Stat(data.PreferencesFile)
	var preferences = data.GetNewPreferences()
	if err == nil {
		log.Println("Preferences file exists, loading preferences...")
		if preferences.Load() != nil {
			log.Fatal("Error occurred while loading preferences" + err.Error())
		}
	} else if os.IsNotExist(err) {
		log.Println("There are no saved preferences, loading default preferences...")
	} else {
		log.Fatal("Error occurred while checking for preferences" + err.Error())
	}
	http.HandleFunc("/images/", handler.ImageHandler)
	apiHandler := handler.NewHandler(preferences, &http.Client{})
	http.HandleFunc("/devices", apiHandler.DevicesHandler)
	http.HandleFunc("/preferences", apiHandler.PreferencesHandler)
	http.HandleFunc("/upload", apiHandler.Upload)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

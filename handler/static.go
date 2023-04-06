package handler

import "net/http"

// ImageHandler is the method used to handle get request for images
func ImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Get the image file name from the URL
		imgName := r.URL.Path[len("/images/"):]
		// Construct the file path to the image
		filePath := "images/" + imgName

		// Serve the image file using http.ServeFile()
		http.ServeFile(w, r, filePath)
	} else {
		// Handling error for other http methods
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

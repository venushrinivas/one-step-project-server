package handler

import (
	"io"
	"main/data"
	"mime/multipart"
	"net/http"
	"os"
)

// OneStepDeviceApiUrl Constant to store the API url
const OneStepDeviceApiUrl = "https://track.onestepgps.com/v3/api/public/device?latest_point=true&api-key=%s"

// DefaultImagePath The url for the default image
const DefaultImagePath = "https://cdn4.iconfinder.com/data/icons/BRILLIANT/transportation/png/400/muscle_car.png"

// Device Structure that holds the required fields for a device
type Device struct {
	DeviceID          string `json:"device_id"`
	DisplayName       string `json:"display_name"`
	ActiveState       string `json:"active_state"`
	Online            bool   `json:"online"`
	Image             string `json:"image"`
	LatestDevicePoint struct {
		Lat          float64 `json:"lat"`
		Lng          float64 `json:"lng"`
		Altitude     float64 `json:"altitude"`
		DeviceStatus struct {
			DriveStatus string `json:"drive_status"`
		} `json:"device_state"`
	} `json:"latest_accurate_device_point"`
}

// ApiResponse Structure to hold the deserialized one step api response. Stores a list of Devices
type ApiResponse struct {
	Devices []Device `json:"result_list"`
}

// Response Get API response structure
type Response struct {
	Message string `json:"message"`
}

// Handler Structure which stores information required for api handler.
// Preferences is the preferences object
// httpClient is the client for making http request
// FileSystem is a wrapper for the os file system
type Handler struct {
	Preferences data.Preferences
	httpClient  *http.Client
	FileSystem  FileSystemInterface
}

// FileSystemInterface which has methods for file operations
type FileSystemInterface interface {
	// MkdirAll creates directories which are needed to create the file
	MkdirAll(path string, perm os.FileMode) error
	// Create Cereates a file with the given name
	Create(name string) (*os.File, error)
	// Copy copies the src file to the destination file
	Copy(dst *os.File, src multipart.File) (written int64, err error)
}

// FileSystem Implements the FileSystemInterface
type FileSystem struct {
}

func (r *FileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (r *FileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (r *FileSystem) Copy(dst *os.File, src multipart.File) (written int64, err error) {
	return io.Copy(dst, src)
}

// NewHandler Function to create a new api handler. accepts a Preferences p, http.Client client and a FileSystemInterface
func NewHandler(p data.Preferences, client *http.Client, fileSystem FileSystemInterface) *Handler {
	return &Handler{Preferences: p, httpClient: client, FileSystem: fileSystem}
}

// enableCors Method to enable cors for a request
func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

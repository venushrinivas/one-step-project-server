package handler

import (
	"io"
	"main/data"
	"mime/multipart"
	"net/http"
	"os"
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
	FileSystem  FileSystemInterface
}

type FileSystemInterface interface {
	MkdirAll(path string, perm os.FileMode) error
	Create(name string) (*os.File, error)
	Copy(dst *os.File, src multipart.File) (written int64, err error)
}

type FileSystem struct {
}

func (r *FileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (r *FileSystem) Create(name string) (*os.File, error) {
	return &os.File{}, nil
}

func (r *FileSystem) Copy(dst *os.File, src multipart.File) (written int64, err error) {
	return io.Copy(dst, src)
}

func NewHandler(p data.Preferences, client *http.Client, fileSystem FileSystemInterface) *Handler {
	return &Handler{Preferences: p, httpClient: client, FileSystem: fileSystem}
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

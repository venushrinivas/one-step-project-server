package test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"main/data"
	"main/handler"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type MockPreferences struct {
	SortColumn        string                   `json:"sort_column"`
	Ascending         bool                     `json:"ascending"`
	NumberOfRows      int                      `json:"number_of_rows"`
	DevicePreferences []data.DevicePreferences `json:"device_preferences"`
}

var testPreferences *MockPreferences

func (preferences *MockPreferences) Load() error {
	return nil
}

func (preferences *MockPreferences) Save() error {
	testPreferences = preferences
	return nil
}

func (preferences *MockPreferences) GetDevicePreferences() []data.DevicePreferences {
	return preferences.DevicePreferences
}

func (preferences *MockPreferences) GetSortColumn() string {
	return preferences.SortColumn
}

func (preferences *MockPreferences) IsAscending() bool {
	return preferences.Ascending
}

func (preferences *MockPreferences) GetNumberOfRows() int {
	return preferences.NumberOfRows
}

func (preferences *MockPreferences) SetDevicePreferences(devicePreferences []data.DevicePreferences) error {
	preferences.DevicePreferences = devicePreferences
	return nil
}

func GetNewPreferences() *MockPreferences {
	return &MockPreferences{NumberOfRows: -1, SortColumn: "display_name", Ascending: true, DevicePreferences: []data.DevicePreferences{}}
}

func TestPreferencesHandler_POST(t *testing.T) {
	var preferences = GetNewPreferences()
	apiHandler := handler.NewHandler(preferences, &http.Client{}, nil)
	var jsonStr = []byte(`{"sort_column":"lat","ascending":false,"number_of_rows":10,"device_preferences":[{"device_id":"abc123","display_name":"Device 1","hidden":false,"image":""}]}`)

	formBuf := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(formBuf)
	defer multipartWriter.Close()

	dataField, err := multipartWriter.CreateFormField("data")
	dataField.Write(jsonStr)

	req, err := http.NewRequest("POST", "/preferences", formBuf)
	assert.NoError(t, err)
	req.Header.Add("content-type", multipartWriter.FormDataContentType())
	req.Form = map[string][]string{
		"data": {string(jsonStr)},
	}

	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(apiHandler.PreferencesHandler)

	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var response handler.Response
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Request processed successfully", response.Message)

	assert.NoError(t, err)
	assert.Equal(t, "Device 1", testPreferences.GetDevicePreferences()[0].DisplayName)
	assert.Equal(t, 1, len(testPreferences.GetDevicePreferences()))
	assert.Equal(t, false, testPreferences.IsAscending())
	assert.Equal(t, 10, testPreferences.GetNumberOfRows())
}

// Testing by giving invalid json
func TestPreferencesHandlerError_POST(t *testing.T) {
	var preferences = GetNewPreferences()
	apiHandler := handler.NewHandler(preferences, &http.Client{}, nil)
	var jsonStr = []byte(`{"sort_column":"lat","ascending":false,"number_of_rows":10,"device_preferences":[{"device_id":"abc123","display_name":"Device 1","hidden":false,"image":""]}`)

	formBuf := new(bytes.Buffer)
	multipartWriter := multipart.NewWriter(formBuf)
	defer multipartWriter.Close()

	dataField, err := multipartWriter.CreateFormField("data")
	dataField.Write(jsonStr)

	req, err := http.NewRequest("POST", "/preferences", formBuf)
	assert.NoError(t, err)
	req.Header.Add("content-type", multipartWriter.FormDataContentType())
	req.Form = map[string][]string{
		"data": {string(jsonStr)},
	}

	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(apiHandler.PreferencesHandler)

	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// Testing invalid methods
func TestPreferencesHandlerError_InvalidMethods(t *testing.T) {
	formBuf := new(bytes.Buffer)
	var preferences = GetNewPreferences()
	apiHandler := handler.NewHandler(preferences, &http.Client{}, nil)
	req, _ := http.NewRequest("PUT", "/preferences", formBuf)
	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(apiHandler.PreferencesHandler)
	handlerFunc.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	req, _ = http.NewRequest("PATCH", "/preferences", formBuf)
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	req, _ = http.NewRequest("DELETE", "/preferences", formBuf)
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestPreferencesHandler_GET(t *testing.T) {
	expected := []byte(`{"result_list":[{"device_id":"1","active_state":"active","display_name":"Test 1","online":true,"device_state":{"drive_status":"off"},"latest_accurate_device_point":{"lat":34.1611778,"lng":-118.1420194,"altitude":254.58}}]}`)
	mockClient := &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader(expected)),
				Header:     make(http.Header),
			}
		}),
	}
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: 5, Ascending: true, SortColumn: "display_name", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	apiHandler := handler.NewHandler(preferences, mockClient, nil)
	formBuf := new(bytes.Buffer)
	req, _ := http.NewRequest("GET", "/preferences", formBuf)
	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(apiHandler.PreferencesHandler)
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Equal(t, `{"sort_column":"display_name","ascending":true,"number_of_rows":5,"device_preferences":[{"device_id":"1","display_name":"Test 1","hidden":false,"image":""}]}`+"\n", rr.Body.String())

}

type FileSystemMock struct {
}

func (r *FileSystemMock) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (r *FileSystemMock) Create(name string) (*os.File, error) {
	serverFileName = "/" + name
	return &os.File{}, nil
}

var serverFileName string
var uploadedFileContent string

func (r *FileSystemMock) Copy(dst *os.File, src multipart.File) (written int64, err error) {
	Buf, err := io.ReadAll(src)
	uploadedFileContent = string(Buf)
	return 0, nil
}

func TestUpload_POST(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: 5, Ascending: true, SortColumn: "display_name", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	apiHandler := handler.NewHandler(preferences, nil, &FileSystemMock{})
	handlerFunc := http.HandlerFunc(apiHandler.Upload)

	file, err := os.Open("images/default.png")

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	_, err = io.Copy(part, file)
	file.Close()
	if err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	writer.Close()

	// Create a new POST request to the test server with the mock request body
	req, err := http.NewRequest("POST", "/upload?device_id=1", body)

	if err != nil {
		t.Fatalf("failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Decode the response body into a Response struct
	var resp handler.Response
	err = json.NewDecoder(rr.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	// Check the response message
	expectedMessage := "/images/1.png"
	assert.Equal(t, expectedMessage, resp.Message)

	expectedFileContent, err := os.ReadFile("images/default.png")

	assert.Equal(t, string(expectedFileContent), uploadedFileContent)
	assert.Equal(t, expectedMessage, serverFileName)
}

func TestImage_GET(t *testing.T) {
	formBuf := new(bytes.Buffer)
	req, _ := http.NewRequest("GET", "/images/default.png", formBuf)
	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(handler.ImageHandler)
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "image/png", rr.Header().Get("Content-Type"))

	expectedImage, err := os.ReadFile("images/default.png")
	if err != nil {
		t.Fatalf(err.Error())
	}
	actualContent := rr.Body.Bytes()

	assert.Equal(t, expectedImage, actualContent)
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

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

// MockPreferences for testing purpose
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

// GetNewPreferences function to return a new mock preferences object with default data
func GetNewPreferences() *MockPreferences {
	return &MockPreferences{NumberOfRows: -1, SortColumn: "display_name", Ascending: true, DevicePreferences: []data.DevicePreferences{}}
}

// TestPreferencesHandler_POST function to test the POST request of preferences api
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

// TestPreferencesHandler_GET function to test the GET request of preferences api
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

// Mocking the file system for test
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

// Test for the upload api
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

// Test for get image api
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

// test for other methods of image api
func TestImage_OtherMethods(t *testing.T) {
	formBuf := new(bytes.Buffer)
	req, _ := http.NewRequest("POST", "/images/default.png", formBuf)
	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(handler.ImageHandler)
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	req, _ = http.NewRequest("PUT", "/images/default.png", formBuf)
	rr = httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	req, _ = http.NewRequest("PATCH", "/images/default.png", formBuf)
	rr = httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	req, _ = http.NewRequest("DELETE", "/images/default.png", formBuf)
	rr = httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

// Test device api get with pagination
func TestDevicesHandlerPagination_GET(t *testing.T) {
	expected, err := os.ReadFile("api_response.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
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
	preferences := &MockPreferences{NumberOfRows: 5, Ascending: true, SortColumn: "display_name", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	apiHandler := handler.NewHandler(preferences, mockClient, nil)
	formBuf := new(bytes.Buffer)
	req, _ := http.NewRequest("GET", "/devices?page=1", formBuf)
	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(apiHandler.DevicesHandler)
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	expectedBytes, err := os.ReadFile("device_response_1.json")
	dst := &bytes.Buffer{}

	json.Compact(dst, expectedBytes)

	assert.Equal(t, dst.String()+"\n", rr.Body.String())

	req, _ = http.NewRequest("GET", "/devices?page=2", formBuf)
	rr = httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	expectedBytes, err = os.ReadFile("device_response_2.json")
	dst = &bytes.Buffer{}

	json.Compact(dst, expectedBytes)

	assert.Equal(t, dst.String()+"\n", rr.Body.String())
}

// Test get method of device api without pagination
func TestDevicesHandlerAllRows_GET(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: -1, Ascending: true, SortColumn: "display_name", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	testDevicesHelper(t, preferences, "device_response_3.json")
}

// Test get method of device api with data sorted by display name
func TestDevicesHandler_SortDisplayName(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: -1, Ascending: true, SortColumn: "display_name", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	testDevicesHelper(t, preferences, "device_response_3.json")

	preferences.Ascending = false
	testDevicesHelper(t, preferences, "device_response_name_descending.json")
}

// Test get method of device api with data sorted by device id
func TestDevicesHandler_SortDeviceID(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: -1, Ascending: true, SortColumn: "device_id", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	testDevicesHelper(t, preferences, "device_response_deviceid_asc.json")

	preferences.Ascending = false
	testDevicesHelper(t, preferences, "device_response_deviceid_dsc.json")
}

// Test get method of device api with data sorted by active state
func TestDevicesHandler_SortActiveState(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: -1, Ascending: true, SortColumn: "active_state", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	testDevicesHelper(t, preferences, "device_response_active_state_asc.json")

	preferences.Ascending = false
	testDevicesHelper(t, preferences, "device_response_active_state_dsc.json")
}

// Test get method of device api with data sorted by online field
func TestDevicesHandler_SortOnline(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: -1, Ascending: true, SortColumn: "online", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	testDevicesHelper(t, preferences, "device_response_online_asc.json")

	preferences.Ascending = false
	testDevicesHelper(t, preferences, "device_response_online_dsc.json")
}

// Test get method of device api with data sorted by latitude
func TestDevicesHandler_SortLatitude(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: -1, Ascending: true, SortColumn: "lat", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	testDevicesHelper(t, preferences, "device_response_lat_asc.json")

	preferences.Ascending = false
	testDevicesHelper(t, preferences, "device_response_lat_dsc.json")
}

// Test get method of device api with data sorted by longitude
func TestDevicesHandler_SortLongitude(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: -1, Ascending: true, SortColumn: "lng", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	testDevicesHelper(t, preferences, "device_response_lng_asc.json")

	preferences.Ascending = false
	testDevicesHelper(t, preferences, "device_response_lng_dsc.json")
}

// Test get method of device api with data sorted by altitude
func TestDevicesHandler_SortAltitude(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: -1, Ascending: true, SortColumn: "altitude", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	testDevicesHelper(t, preferences, "device_response_altitude_asc.json")

	preferences.Ascending = false
	testDevicesHelper(t, preferences, "device_response_altitude_dsc.json")
}

// Test get method of device api with data sorted by drive status
func TestDevicesHandler_SortDriveStatus(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: -1, Ascending: true, SortColumn: "drive_status", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	testDevicesHelper(t, preferences, "device_response_drive_status_asc.json")

	preferences.Ascending = false
	testDevicesHelper(t, preferences, "device_response_drive_status_dsc.json")
}

// Testing device api for other http methods
func TestDevicesHandler_OtherMethods(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: 5, Ascending: true, SortColumn: "display_name", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	apiHandler := handler.NewHandler(preferences, nil, nil)
	formBuf := new(bytes.Buffer)
	handlerFunc := http.HandlerFunc(apiHandler.DevicesHandler)
	req, _ := http.NewRequest("POST", "/devices?page=2", formBuf)
	rr := httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	req, _ = http.NewRequest("PUT", "/devices?page=2", formBuf)
	rr = httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	req, _ = http.NewRequest("PATCH", "/devices?page=2", formBuf)
	rr = httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	req, _ = http.NewRequest("DELETE", "/devices?page=2", formBuf)
	rr = httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

// Testing get devices api with invalid pages
func TestDevicesHandlerInvalidPages_GET(t *testing.T) {
	var devicePreferences []data.DevicePreferences
	preferences := &MockPreferences{NumberOfRows: 5, Ascending: true, SortColumn: "display_name", DevicePreferences: append(devicePreferences, data.DevicePreferences{Image: "images/default.png", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})}
	expected, err := os.ReadFile("api_response.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	mockClient := &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader(expected)),
				Header:     make(http.Header),
			}
		}),
	}
	apiHandler := handler.NewHandler(preferences, mockClient, nil)
	formBuf := new(bytes.Buffer)
	handlerFunc := http.HandlerFunc(apiHandler.DevicesHandler)
	req, _ := http.NewRequest("GET", "/devices?page=3", formBuf)
	rr := httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "Page does not exist\n", rr.Body.String())

	req, _ = http.NewRequest("GET", "/devices?page=0", formBuf)
	rr = httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "Page does not exist\n", rr.Body.String())
}

// Helper method for devices api
func testDevicesHelper(t *testing.T, preferences data.Preferences, expectedJson string) {
	t.Helper()
	expected, err := os.ReadFile("api_response.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	mockClient := &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader(expected)),
				Header:     make(http.Header),
			}
		}),
	}
	apiHandler := handler.NewHandler(preferences, mockClient, nil)

	handlerFunc := http.HandlerFunc(apiHandler.DevicesHandler)
	formBuf := new(bytes.Buffer)
	req, _ := http.NewRequest("GET", "/devices?page=1", formBuf)
	rr := httptest.NewRecorder()
	handlerFunc.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	expectedBytes, err := os.ReadFile(expectedJson)
	dst := &bytes.Buffer{}

	json.Compact(dst, expectedBytes)

	assert.Equal(t, dst.String()+"\n", rr.Body.String())
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

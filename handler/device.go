package handler

import (
	"encoding/json"
	"fmt"
	"main/data"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

// SortDevice structure used for sorting devices
type SortDevice struct {
	devices     []Device
	preferences data.Preferences
}

// GetDevicesResponse structure representing the data for the devices get api
type GetDevicesResponse struct {
	Devices      []Device `json:"devices"`
	PageNumber   int      `json:"page_number"`
	NextPage     bool     `json:"next_page"`
	PreviousPage bool     `json:"previous_page"`
}

// Len returns the size of the devices array in SortDevice
func (sortDevice SortDevice) Len() int {
	return len(sortDevice.devices)
}

// Swap swaps the devices in position i with device in position j
func (sortDevice SortDevice) Swap(i, j int) {
	sortDevice.devices[i], sortDevice.devices[j] = sortDevice.devices[j], sortDevice.devices[i]
}

// Less returns true if device at position i is lesser than device at position j based on sort column. returns false otherwise
func (sortDevice SortDevice) Less(i, j int) bool {
	preferences := sortDevice.preferences
	devices := sortDevice.devices
	switch preferences.GetSortColumn() {
	case "device_id":
		if preferences.IsAscending() {
			return strings.Compare(devices[i].DeviceID, devices[j].DeviceID) < 0
		} else {
			return strings.Compare(devices[i].DeviceID, devices[j].DeviceID) > 0
		}
	case "display_name":
		if preferences.IsAscending() {
			return strings.Compare(strings.ToLower(devices[i].DisplayName), strings.ToLower(devices[j].DisplayName)) < 0
		} else {
			return strings.Compare(strings.ToLower(devices[i].DisplayName), strings.ToLower(devices[j].DisplayName)) > 0
		}
	case "active_state":
		if preferences.IsAscending() {
			return strings.Compare(devices[i].ActiveState, devices[j].ActiveState) < 0
		} else {
			return strings.Compare(devices[i].ActiveState, devices[j].ActiveState) > 0
		}
	case "online":
		if !preferences.IsAscending() {
			return devices[i].Online && !devices[j].Online
		} else {
			return devices[j].Online && !devices[i].Online
		}
	case "lat":
		if preferences.IsAscending() {
			return devices[i].LatestDevicePoint.Lat < devices[j].LatestDevicePoint.Lat
		} else {
			return devices[i].LatestDevicePoint.Lat > devices[j].LatestDevicePoint.Lat
		}
	case "lng":
		if preferences.IsAscending() {
			return devices[i].LatestDevicePoint.Lng < devices[j].LatestDevicePoint.Lng
		} else {
			return devices[i].LatestDevicePoint.Lng > devices[j].LatestDevicePoint.Lng
		}
	case "altitude":
		if preferences.IsAscending() {
			return devices[i].LatestDevicePoint.Altitude < devices[j].LatestDevicePoint.Altitude
		} else {
			return devices[i].LatestDevicePoint.Altitude > devices[j].LatestDevicePoint.Altitude
		}
	case "drive_status":
		if preferences.IsAscending() {
			return strings.Compare(devices[i].LatestDevicePoint.DeviceStatus.DriveStatus, devices[j].LatestDevicePoint.DeviceStatus.DriveStatus) < 0
		} else {
			return strings.Compare(devices[i].LatestDevicePoint.DeviceStatus.DriveStatus, devices[j].LatestDevicePoint.DeviceStatus.DriveStatus) > 0
		}
	}
	return false
}

// sortDevices helper method for sorting the devices based on user preferences
func sortDevices(devices []Device, preferences data.Preferences) []Device {
	sortDevice := SortDevice{
		devices:     devices,
		preferences: preferences,
	}
	sort.Sort(sortDevice)
	return sortDevice.devices
}

// DevicesHandler handler method for the get request for the devices api. Accepts a request and response object.
func (h *Handler) DevicesHandler(w http.ResponseWriter, r *http.Request) {
	preferences := h.Preferences
	enableCors(w)
	if r.Method == http.MethodGet {
		// Constructing the api url by appending the api key
		apiUrl := fmt.Sprintf(OneStepDeviceApiUrl, os.Getenv("API_KEY"))
		res, err := h.httpClient.Get(apiUrl)

		// Checking if the api call had an error, and if it has sending the error in response
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Extracting the page number query param from the url
		queryParams := r.URL.Query()
		page, err := strconv.Atoi(queryParams.Get("page"))

		if err != nil {
			page = 1
		}
		if page == 0 {
			http.Error(w, "Page does not exist", http.StatusBadRequest)
			return
		}
		defer res.Body.Close()
		var resultList ApiResponse

		// Decoding the response and deserializing into the resultList object
		err = json.NewDecoder(res.Body).Decode(&resultList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var visibleDevices []Device

		// Appending the individual device preferences to the response
		if preferences.GetDevicePreferences() != nil {
			visibleDevices = make([]Device, 0)
			for _, device := range resultList.Devices {
				matched := false
				for _, devicePreference := range preferences.GetDevicePreferences() {
					if device.DeviceID == devicePreference.DeviceID {
						matched = true
						device.Image = devicePreference.Image
						if !devicePreference.Hidden {
							visibleDevices = append(visibleDevices, device)
						}
					}
				}
				if !matched {
					device.Image = DefaultImagePath
					visibleDevices = append(visibleDevices, device)
				}
			}
		} else {
			visibleDevices = resultList.Devices
		}
		var devicesResponse GetDevicesResponse
		devicesResponse.PageNumber = page

		// Sorting the devices based on the user preferences
		visibleDevices = sortDevices(visibleDevices, preferences)

		// Handling pagination
		if preferences.GetNumberOfRows() != -1 {
			if (page-1)*preferences.GetNumberOfRows() >= len(visibleDevices) {
				http.Error(w, "Page does not exist", http.StatusBadRequest)
				return
			}
			if (page)*preferences.GetNumberOfRows() >= len(visibleDevices) {
				devicesResponse.NextPage = false
			} else {
				devicesResponse.NextPage = true
			}
			visibleDevices = visibleDevices[(page-1)*preferences.GetNumberOfRows() : int(math.Min(float64((page)*preferences.GetNumberOfRows()), float64(len(visibleDevices))))]
			if page == 1 {
				devicesResponse.PreviousPage = false
			} else {
				devicesResponse.PreviousPage = true
			}
		} else {
			devicesResponse.PreviousPage = false
			devicesResponse.NextPage = false
		}
		devicesResponse.Devices = visibleDevices
		// Encoding the response to json format for response
		json.NewEncoder(w).Encode(devicesResponse)
	} else {
		// Handling error for all other http methods
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

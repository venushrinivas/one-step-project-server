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

type SortDevice struct {
	devices     []Device
	preferences data.Preferences
}

type GetDevicesResponse struct {
	Devices      []Device `json:"devices"`
	PageNumber   int      `json:"page_number"`
	NextPage     bool     `json:"next_page"`
	PreviousPage bool     `json:"previous_page"`
}

func (sortDevice SortDevice) Len() int {
	return len(sortDevice.devices)
}

func (sortDevice SortDevice) Swap(i, j int) {
	sortDevice.devices[i], sortDevice.devices[j] = sortDevice.devices[j], sortDevice.devices[i]
}

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

func sortDevices(devices []Device, preferences data.Preferences) []Device {
	sortDevice := SortDevice{
		devices:     devices,
		preferences: preferences,
	}
	sort.Sort(sortDevice)
	return sortDevice.devices
}

func (h *Handler) DevicesHandler(w http.ResponseWriter, r *http.Request) {
	preferences := h.Preferences
	enableCors(w)
	if r.Method == http.MethodGet {
		apiUrl := fmt.Sprintf(OneStepDeviceApiUrl, os.Getenv("API_KEY"))
		res, err := h.httpClient.Get(apiUrl)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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

		err = json.NewDecoder(res.Body).Decode(&resultList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var visibleDevices []Device
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
					device.Image = "/images/default.png"
					visibleDevices = append(visibleDevices, device)
				}
			}
		} else {
			visibleDevices = resultList.Devices
		}
		var devicesResponse GetDevicesResponse
		devicesResponse.PageNumber = page

		visibleDevices = sortDevices(visibleDevices, preferences)

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
		json.NewEncoder(w).Encode(devicesResponse)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

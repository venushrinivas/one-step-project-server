package data

import (
	"encoding/json"
	"os"
)

type DevicePreferences struct {
	DeviceID    string `json:"device_id"`
	DisplayName string `json:"display_name"`
	Hidden      bool   `json:"hidden"`
	Image       string `json:"image"`
}

type Preferences interface {
	Load() error
	Save() error
	GetDevicePreferences() []DevicePreferences
	SetDevicePreferences(devicePreferences []DevicePreferences) error
	GetSortColumn() string
	IsAscending() bool
	GetNumberOfRows() int
}

type PreferencesImpl struct {
	SortColumn        string              `json:"sort_column"`
	Ascending         bool                `json:"ascending"`
	NumberOfRows      int                 `json:"number_of_rows"`
	DevicePreferences []DevicePreferences `json:"device_preferences"`
}

const PreferencesFile = "preferences.json"

func GetNewPreferences() *PreferencesImpl {
	return &PreferencesImpl{NumberOfRows: -1, SortColumn: "device_name", Ascending: true, DevicePreferences: []DevicePreferences{}}
}

func (preferences *PreferencesImpl) Load() error {
	file, err := os.Open("preferences.json")
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(preferences)

	return err
}

func (preferences *PreferencesImpl) Save() error {
	file, err := os.Create(PreferencesFile)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(preferences)
	return err
}

func (preferences *PreferencesImpl) GetDevicePreferences() []DevicePreferences {
	return preferences.DevicePreferences
}

func (preferences *PreferencesImpl) GetSortColumn() string {
	return preferences.SortColumn
}

func (preferences *PreferencesImpl) IsAscending() bool {
	return preferences.Ascending
}

func (preferences *PreferencesImpl) GetNumberOfRows() int {
	return preferences.NumberOfRows
}

func (preferences *PreferencesImpl) SetDevicePreferences(devicePreferences []DevicePreferences) error {
	preferences.DevicePreferences = devicePreferences
	return preferences.Save()
}

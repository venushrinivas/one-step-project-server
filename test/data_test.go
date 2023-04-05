package test

import (
	"github.com/stretchr/testify/assert"
	"main/data"
	"os"
	"testing"
)

func TestDevicePreferences(t *testing.T) {
	preferences := data.GetNewPreferences()
	var devicePreferences []data.DevicePreferences
	devicePreferences = append(devicePreferences, data.DevicePreferences{Image: "", DeviceID: "1", Hidden: false, DisplayName: "Test 1"})
	_, err := os.Stat(data.PreferencesFile)
	assert.True(t, os.IsNotExist(err))
	preferences.SetDevicePreferences(devicePreferences)

	_, err = os.Stat(data.PreferencesFile)

	assert.False(t, os.IsNotExist(err))

	preferences = data.GetNewPreferences()
	assert.Equal(t, 0, len(preferences.GetDevicePreferences()))
	preferences.Load()

	assert.Equal(t, true, preferences.IsAscending())
	assert.Equal(t, "display_name", preferences.GetSortColumn())
	assert.Equal(t, -1, preferences.GetNumberOfRows())
	assert.Equal(t, 1, len(preferences.GetDevicePreferences()))

	os.Remove(data.PreferencesFile)
}

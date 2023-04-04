package twomes

import "time"

// An Upload is a collection of measurements, with additional information.
type Upload struct {
	ID           uint          `json:"id"`
	DeviceID     uint          `json:"device_id"`
	ServerTime   Time          `json:"server_time"`
	DeviceTime   Time          `json:"device_time"`
	Size         int           `json:"size"`
	Measurements []Measurement `json:"measurements,omitempty"`
}

// Create a new Upload.
func MakeUpload(deviceID uint, deviceTime Time, measurements []Measurement) Upload {
	return Upload{
		DeviceID:     deviceID,
		ServerTime:   Time(time.Now().UTC()),
		DeviceTime:   deviceTime,
		Size:         len(measurements),
		Measurements: measurements,
	}
}

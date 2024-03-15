package mppt

import (
	"encoding/binary"
	"fmt"
)

// MPPTData contains the raw values from a decrypted advertisement ciphertext
type MPPTData struct {
	DeviceState    byte
	ChargerError   byte
	BatteryVoltage []byte
	BatteryCurrent []byte
	YieldToday     []byte
	PVPower        []byte
	LoadCurrent    []byte // 9 bits, so we store it in 3 byte, i guess? doesnt really matter, dont care about this metric
}

// ReadableData contains the most relevant data in human readable form
type ReadableData struct {
	BatteryVoltage float32 // V
	BatteryCurrent float32 // A
	YieldToday     float32 // kWH
	PVPower        uint16  // W
}

// ExtractReadableData parses the raw data to sensible values
func (m *MPPTData) ExtractReadableData() ReadableData {
	r := ReadableData{
		BatteryVoltage: float32(binary.LittleEndian.Uint16(m.BatteryVoltage)) * 0.01,
		BatteryCurrent: float32(binary.LittleEndian.Uint16(m.BatteryCurrent)) * 0.1,
		YieldToday:     float32(binary.LittleEndian.Uint16(m.YieldToday)) * 0.01,
		PVPower:        binary.LittleEndian.Uint16(m.PVPower),
	}
	return r
}

// ParseDecrypted parses the decrypted ciphertext to MPPT DataPoints
func ParseDecrypted(plaintext []byte) (*MPPTData, error) {
	if len(plaintext) < 12 {
		return nil, fmt.Errorf("decoded plaintext too short! should be at least 12 bytes, but is %d", len(plaintext))
	}
	mppt := MPPTData{
		DeviceState:    plaintext[0],
		ChargerError:   plaintext[1],
		BatteryVoltage: plaintext[2:4],
		BatteryCurrent: plaintext[4:6],
		YieldToday:     plaintext[6:8],
		PVPower:        plaintext[8:10],
		LoadCurrent:    plaintext[10:],
	}

	return &mppt, nil
}

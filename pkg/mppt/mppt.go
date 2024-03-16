package mppt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

// MPPTConnection looks for a BLE device with the given UUID and sends a message to the AdverisementChan for each advertisement
type MPPTConnection struct {
	AdvertisementChan *chan map[uint16][]byte
	UUID              string
	Timeout           time.Duration
	Key               string
}

// MPPTData contains the most relevant data in human readable form
type MPPTData struct {
	DeviceState    uint8   // custom victron state values. 3 == MPPT
	ChargerError   uint8   // custom victron error codes.
	BatteryVoltage float32 // V
	BatteryCurrent float32 // A
	YieldToday     float32 // kWH
	PVPower        uint16  // W
	LoadCurrent    float32 // W TODO:FIXME: currently unable to parse the 9 bits correctly
}

// New create a new MPPTConnection for receiving and decrypting victron mppt data
func NewMPPT(victronUUID, victronKey string) (*MPPTConnection, error) {
	advChan := make(chan map[uint16][]byte)
	err := adapter.Enable()
	if err != nil {
		return nil, err
	}
	vc := MPPTConnection{UUID: victronUUID, AdvertisementChan: &advChan, Timeout: 1 * time.Second, Key: victronKey}

	return &vc, nil
}

// StartScanning starts looking for BLE Advertisement from the correct UUID and sends it to the channel
func (vc *MPPTConnection) StartScanning() error {
	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		log.Trace().Msg("scanning...")
		if device.Address.String() == vc.UUID {
			manufacturerData := device.ManufacturerData()
			(*vc.AdvertisementChan) <- manufacturerData
			log.Debug().Msgf("found advertisement from victron: %s %v %s", device.Address.String(), device.RSSI, device.LocalName())
			// stop scanning after receiving advertisement. Restart after Timeout has passed
			adapter.StopScan()
			go func() {
				time.Sleep(vc.Timeout)
				vc.StartScanning()
			}()
		}
	})
	if err != nil {
		return err
	}
	return nil
}

func (vc *MPPTConnection) GetChannel() *chan map[uint16][]byte {
	return vc.AdvertisementChan
}

// Decrypt takes the nonce from the payload, and the Key from MPPTConnection to AES-CRT decrypt a received advertisement
func (vc *MPPTConnection) Decrypt(vci *victronAdvertisementInfo) ([]byte, error) {
	key, err := hex.DecodeString(vc.Key)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Got key (%d bytes): %x", len(key)*8, key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("blocksize: %d", aes.BlockSize)

	plaintext := make([]byte, len(vci.Ciphertext))
	// create a new empty 128 bit IV. We will put our LSB nonce at the end
	iv := make([]byte, 16)
	iv[0] = vci.Nonce[1]
	iv[1] = vci.Nonce[0]

	// normal AES CRT decryption
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, vci.Ciphertext)
	return plaintext, nil
}

// Parse parses the unencrypted values from a received advertisement
func (vc *MPPTConnection) Parse(ciphertext []byte) (*MPPTData, error) {
	log.Debug().Msgf("raw adv data: %x", ciphertext)
	if len(ciphertext) < 20 { // is 20, but i thought it should be 24? weird
		return nil, fmt.Errorf("ciphertext too short. Should be at least 24 byte. Was: %d", len(ciphertext))
	}
	// nonce is least significant bit first
	nonceReader := bytes.NewReader([]byte{ciphertext[6], ciphertext[5]})
	var nonce []byte = make([]byte, 2)
	err := binary.Read(nonceReader, binary.LittleEndian, &nonce) // dont think we actually need this, but it should be LSB
	if err != nil {
		return nil, err
	}
	// first 4 bytes are useless
	vai := victronAdvertisementInfo{
		RecordType:               ciphertext[4],
		Nonce:                    nonce,
		FirstByteOfEncryptionKey: ciphertext[7],
		Ciphertext:               ciphertext[8:],
	}

	plaintext, err := vc.Decrypt(&vai)
	if err != nil {
		return nil, err
	}
	mpptData, err := parseDecrypted(plaintext)
	if err != nil {
		return nil, err
	}
	return mpptData, nil
}

// ParseDecrypted parses the decrypted ciphertext to MPPT DataPoints
func parseDecrypted(plaintext []byte) (*MPPTData, error) {
	if len(plaintext) < 12 {
		return nil, fmt.Errorf("decoded plaintext too short! should be at least 12 bytes, but is %d", len(plaintext))
	}
	raw := rawData{
		DeviceState:    plaintext[0],
		ChargerError:   plaintext[1],
		BatteryVoltage: plaintext[2:4],
		BatteryCurrent: plaintext[4:6],
		YieldToday:     plaintext[6:8],
		PVPower:        plaintext[8:10],
		LoadCurrent:    plaintext[10:],
	}

	log.Debug().Msgf("raw deviceState: %x, raw chargerError: %x, raw load current: %x", raw.DeviceState, raw.ChargerError, raw.LoadCurrent)

	mppt := MPPTData{
		DeviceState:    uint8(raw.DeviceState),
		ChargerError:   uint8(raw.DeviceState),
		BatteryVoltage: float32(binary.LittleEndian.Uint16(raw.BatteryVoltage)) * 0.01,
		BatteryCurrent: float32(binary.LittleEndian.Uint16(raw.BatteryCurrent)) * 0.1,
		YieldToday:     float32(binary.LittleEndian.Uint16(raw.YieldToday)) * 0.01,
		PVPower:        binary.LittleEndian.Uint16(raw.PVPower),
		LoadCurrent:    float32(binary.LittleEndian.Uint16(raw.LoadCurrent)) * 0.1,
	}
	return &mppt, nil
}

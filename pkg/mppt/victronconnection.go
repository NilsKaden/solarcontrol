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

// VictronConnection looks for a BLE device with the given UUID and sends a message to the AdverisementChan for each advertisement
type VictronConnection struct {
	AdvertisementChan *chan map[uint16][]byte
	UUID              string
	Timeout           time.Duration
	Key               string
}

// VictronAdvertisementInfo contains the raw response, still encrypted
type VictronAdvertisementInfo struct {
	RecordType               byte
	Nonce                    []byte
	FirstByteOfEncryptionKey byte // whatever we shall use this for
	Ciphertext               []byte
}

// New create a new VictronConnection for receiving and decrypting victron mppt data
func New(victronUUID, victronKey string) (*VictronConnection, error) {
	advChan := make(chan map[uint16][]byte)
	err := adapter.Enable()
	if err != nil {
		return nil, err
	}
	vc := VictronConnection{UUID: victronUUID, AdvertisementChan: &advChan, Timeout: 1 * time.Second, Key: victronKey}

	return &vc, nil
}

// StartScanning starts looking for BLE Advertisement from the correct UUID and sends it to the channel
func (vc *VictronConnection) StartScanning() error {
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

// Decrypt takes the nonce from the payload, and the Key from VictronConnection to AES-CRT decrypt a received advertisement
func (vc *VictronConnection) Decrypt(vci *VictronAdvertisementInfo) ([]byte, error) {
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
func Parse(ciphertext []byte) (*VictronAdvertisementInfo, error) {
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
	vic := VictronAdvertisementInfo{
		RecordType:               ciphertext[4],
		Nonce:                    nonce,
		FirstByteOfEncryptionKey: ciphertext[7],
		Ciphertext:               ciphertext[8:],
	}

	return &vic, nil
}

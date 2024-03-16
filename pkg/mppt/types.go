package mppt

// victronAdvertisementInfo contains the raw response, still encrypted
type victronAdvertisementInfo struct {
	RecordType               byte
	Nonce                    []byte
	FirstByteOfEncryptionKey byte // whatever we shall use this for
	Ciphertext               []byte
}

// rawData contains the raw values from a decrypted advertisement ciphertext
type rawData struct {
	DeviceState    byte
	ChargerError   byte
	BatteryVoltage []byte
	BatteryCurrent []byte
	YieldToday     []byte
	PVPower        []byte
	LoadCurrent    []byte // 9 bits, so we store it in 3 byte, i guess? doesnt really matter, dont care about this metric
}

package controller

import (
	"solarcontrol/pkg/ahoy"
	"solarcontrol/pkg/mppt"

	"github.com/rs/zerolog/log"
)

type AhoyInterface interface {
	GetInverterInfo() (*ahoy.InverterInfo, error)
	SetInverterPower(int, bool) error
}

type MPPTInterface interface {
	StartScanning() error
	Parse([]byte) (*mppt.MPPTData, error)
	GetChannel() *chan map[uint16][]byte
}

func Start(ahoy AhoyInterface, mppt MPPTInterface) error {
	ch := mppt.GetChannel()
	go mppt.StartScanning()

	for btAdvertisement := range *ch {
		for _, btAdvertisementBytes := range btAdvertisement { // its a map with one key. We only care about the value
			data, err := mppt.Parse(btAdvertisementBytes)
			if err != nil {
				return (err)
			}

			log.Info().Msgf(
				"VBatt: %.2fV IBatt: %.2fA Pday: %.2fkWh PV: %dW State: %d Error: %d LoadCurrent: %.2fA ",
				data.BatteryVoltage, data.BatteryCurrent, data.YieldToday, data.PVPower, data.DeviceState, data.ChargerError, data.LoadCurrent)

		}
	}

	return nil
}

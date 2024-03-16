package controller

import (
	"github.com/rs/zerolog/log"
	"solarcontrol/pkg/ahoy"
	"solarcontrol/pkg/mppt"
)

// adjust for different battery types
const (
	BatteryTurnOffVoltage = 25.6 // 20%
	BatteryTurnOnVoltage  = 26.7 // about 70% when charging, implement in the future
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

type Controller struct {
	ahoy AhoyInterface
	mppt MPPTInterface
}

func NewController(ahoy AhoyInterface, mppt MPPTInterface) (*Controller, error) {
	ii, err := ahoy.GetInverterInfo()
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("inverter info: %v", ii)

	c := Controller{
		ahoy: ahoy,
		mppt: mppt,
	}

	return &c, nil
}

func (c *Controller) Start() error {
	ch := c.mppt.GetChannel()
	go c.mppt.StartScanning()

	for btAdvertisement := range *ch {
		for _, btAdvertisementBytes := range btAdvertisement { // its a map with one key. We only care about the value
			data, err := c.mppt.Parse(btAdvertisementBytes)
			if err != nil {
				return (err)
			}

			log.Info().Msgf(
				"VBatt: %.2fV IBatt: %.2fA Pday: %.2fkWh PV: %dW State: %d Error: %d LoadCurrent: %.2fA ",
				data.BatteryVoltage, data.BatteryCurrent, data.YieldToday, data.PVPower, data.DeviceState, data.ChargerError, data.LoadCurrent)

			if data.BatteryVoltage < BatteryTurnOffVoltage {
				log.Info().Msgf("turning off inverter at %.2fV", data.BatteryVoltage)
				err := c.ahoy.SetInverterPower(0, true)
				if err != nil {
					log.Error().Err(err).Msgf("UNABLE TO SHUTDOWN INVERTER, BUT BATTERY IS LOW. PANIC")
					// turn off myStrom Smart Plug
				}
			}
		}
	}

	return nil
}

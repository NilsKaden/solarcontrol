package controller

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"solarcontrol/pkg/ahoy"
	"solarcontrol/pkg/mppt"
	"time"
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

type MystromInterface interface {
	Disable() error
}

type Controller struct {
	ahoy           AhoyInterface
	mppt           MPPTInterface
	mystrom        MystromInterface
	shutoffVoltage float32
}

func NewController(ahoy AhoyInterface, mppt MPPTInterface, mystrom MystromInterface, shutOffVoltage float32) (*Controller, error) {
	ii, err := ahoy.GetInverterInfo()
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("inverter info: %#v", ii)

	if len(ii.Ch) > 0 && ii.Ch[1] != nil {
		voltage := ii.Ch[1][0]
		minAcceptableVoltage := shutOffVoltage * 0.8
		maxAcceptableVoltage := shutOffVoltage * 1.2

		// if the current voltage is more than 20% different from the shutoffVoltage, we return an error.
		if voltage < minAcceptableVoltage || voltage > maxAcceptableVoltage {
			return nil, fmt.Errorf("Voltage %.1f measured at ch 1 is not in range of ShutoffVoltage: %.1f", voltage, shutOffVoltage)
		}
	}

	c := Controller{
		ahoy:           ahoy,
		mppt:           mppt,
		mystrom:        mystrom,
		shutoffVoltage: shutOffVoltage,
	}

	return &c, nil
}

func (c *Controller) TurnOffInverterIfVoltageLow(voltage float32) bool {
	var turnedOffInverter bool = false

	if voltage < c.shutoffVoltage {
		log.Info().Msgf("turning off inverter at %.2fV", voltage)
		err := c.ahoy.SetInverterPower(0, false)
		if err != nil {
			log.Error().Err(err).Msgf("UNABLE TO SHUTDOWN INVERTER, BUT BATTERY IS LOW. PANIC")
		}
		turnedOffInverter = true
		if c.mystrom != nil {
			time.Sleep(5 * time.Second)
			// turn off myStrom Smart Plug for extra safety
			c.mystrom.Disable()
		}
	}
	return turnedOffInverter
}

func (c *Controller) Start() error {
	ch := c.mppt.GetChannel()
	go c.mppt.StartScanning()

	for btAdvertisement := range *ch {

		// its a map with one key. We only care about the value
		for _, btAdvertisementBytes := range btAdvertisement {
			data, err := c.mppt.Parse(btAdvertisementBytes)
			if err != nil {
				return (err)
			}

			log.Info().Msgf(
				"VBatt: %.2fV IBatt: %.2fA Pday: %.2fkWh PV: %dW State: %d Error: %d LoadCurrent: %.2fA ",
				data.BatteryVoltage, data.BatteryCurrent, data.YieldToday, data.PVPower, data.DeviceState, data.ChargerError, data.LoadCurrent)

			c.TurnOffInverterIfVoltageLow(data.BatteryVoltage)
		}
	}

	return nil
}

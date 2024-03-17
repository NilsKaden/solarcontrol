package main

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"solarcontrol/pkg/ahoy"
	"solarcontrol/pkg/controller"
	"solarcontrol/pkg/mppt"
	"solarcontrol/pkg/mystrom"
	"strconv"
	"time"
)

type Config struct {
	VictronUUID     string `env:"VICTRON_UUID"`
	VictronKey      string `env:"VICTRON_KEY"`
	InverterID      string `env:"INVERTER_ID" env-default:"0"`
	AhoyEndpoint    string `env:"AHOY_ENDPOINT"`
	ShutoffVoltage  string `env:"SHUTOFF_VOLTAGE"`
	MystromEndpoint string `env:"MYSTROM_ENDPOINT"`
}

var cfg Config

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Warn().Msg("unable to read config")
		panic(err)
	}
	log.Info().Msg("starting")

	ah := ahoy.NewAhoy(cfg.InverterID, cfg.AhoyEndpoint)

	mp, err := mppt.NewMPPT(cfg.VictronUUID, cfg.VictronKey)
	if err != nil {
		panic(err)
	}

	ms := mystrom.NewMystrom(cfg.MystromEndpoint)

	// parse shutoff voltage
	shutoffVoltage, err := strconv.ParseFloat(cfg.ShutoffVoltage, 32)
	if err != nil {
		log.Error().Err(err).Msgf("should be able to parse env var SHUTOFF_VOLTAGE to a valid float")
	}
	if shutoffVoltage < 10 || shutoffVoltage > 30 {
		log.Warn().Msgf("atypical shutoffVoltage supplied. Got %.2f. Ignore if you're running a system below 12V or above 26V", shutoffVoltage)
	}
	if shutoffVoltage < 1 {
		panic(fmt.Errorf("supplied shutoffVoltage off %.1f. You must be joking, right?", shutoffVoltage))
	}

	c, err := controller.NewController(ah, mp, ms, float32(shutoffVoltage))
	if err != nil {
		panic(err)
	}

	for {
		err := c.Start()
		if err != nil {
			log.Error().Err(err).Msg("error scanning. Waiting for a minute and trying again")
			time.Sleep(time.Minute)
		}
	}
}

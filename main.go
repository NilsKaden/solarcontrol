package main

import (
	"solarcontrol/pkg/ahoy"
	"solarcontrol/pkg/controller"
	"solarcontrol/pkg/mppt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	VictronUUID  string `env:"VICTRON_UUID"`
	VictronKey   string `env:"VICTRON_KEY"`
	InverterID   string `env:"INVERTER_ID" env-default:"0"`
	AhoyEndpoint string `env:"AHOY_ENDPOINT"`
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
	ii, err := ah.GetInverterInfo()
	if err != nil {
		panic(err)
	}
	log.Info().Msgf("inverter info: %v", ii)

	vc, err := mppt.New(cfg.VictronUUID, cfg.VictronKey)
	if err != nil {
		panic(err)
	}

	for {
		err := controller.Start(ah, vc)
		if err != nil {
			log.Error().Err(err).Msg("error scanning. Waiting for a minute and trying again")
			time.Sleep(time.Minute)
		}
	}
}

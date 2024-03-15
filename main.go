package main

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"solarcontrol/pkg/mppt"
)

type Config struct {
	VictronUUID string `env:"VICTRON_UUID"`
	VictronKey  string `env:"VICTRON_KEY"`
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

	vc, err := mppt.New(cfg.VictronUUID, cfg.VictronKey)
	if err != nil {
		panic(err)
	}
	go vc.StartScanning()

	for advMsg := range *vc.AdvertisementChan {
		for _, advBytes := range advMsg {
			vci, err := mppt.Parse(advBytes)
			if err != nil {
				panic(err)
			}
			plaintext, err := vc.Decrypt(vci)
			if err != nil {
				panic(err)
			}
			mpptData, err := mppt.ParseDecrypted(plaintext)
			if err != nil {
				return
			}
			rel := mpptData.ExtractReadableData()

			log.Info().Msgf("VBatt: %.2fV IBatt: %.2fA Pday: %.2fkWh Wpv: %dW", rel.BatteryVoltage, rel.BatteryCurrent, rel.YieldToday, rel.PVPower)

		}
	}
}

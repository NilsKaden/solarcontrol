package controller

import "solarcontrol/pkg/ahoy"

type AhoyInterface interface {
	GetInverterInfo() (*ahoy.InverterInfo, error)
	SetInverterPower(int, bool) error
}

type MPPTInterface interface {
	StartScanning()
	Parse()
}

func Start(ahoy AhoyInterface, mppt MPPTInterface) error {
	go mppt.StartScanning()
	return nil
}

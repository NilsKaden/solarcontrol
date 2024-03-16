package ahoy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const InverterRoute = "/api/inverter/id/"
const CtrlRoute = "/api/ctrl"

const CtrlPersistent = "limit_persistent_absolute"
const CtrlNonPersistent = "limit_nonpersistent_absolute"

type Ahoy struct {
	Client       http.Client
	AhoyEndpoint string
	InverterID   string
}

// NewAhoy creates a new ahoy client
func NewAhoy(inverterID, ahoyEndpoint string) *Ahoy {
	a := Ahoy{
		Client:       *http.DefaultClient,
		AhoyEndpoint: ahoyEndpoint,
		InverterID:   inverterID,
	}
	return &a
}

// GetInverterInfo returns a list of all inverters from ahoy DTU api
func (a *Ahoy) GetInverterInfo() (*InverterInfo, error) {
	req, err := http.NewRequest(http.MethodGet, a.AhoyEndpoint+InverterRoute+a.InverterID, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var info InverterInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// SetInverterPower sets the current inverter power in watts. Doesn't apply a
func (a *Ahoy) SetInverterPower(powerLimitWatt int, persistent bool) error {
	id, err := strconv.Atoi(a.InverterID)
	if err != nil {
		return err
	}

	r := CtrlRequest{
		ID:    id,
		Token: "*",
		Cmd:   CtrlNonPersistent,
		Val:   strconv.Itoa(powerLimitWatt),
	}
	if persistent {
		r.Cmd = CtrlPersistent
	}

	reqJson, err := json.Marshal(r)
	if err != nil {
		return err
	}
	reqBody := bytes.NewReader(reqJson)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, a.AhoyEndpoint+CtrlRoute, reqBody)
	if err != nil {
		return err
	}
	resp, err := a.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var ctrlResp CtrlResponse
	err = json.Unmarshal(body, &ctrlResp)
	if err != nil {
		return err
	}

	if ctrlResp.Success != true {
		return fmt.Errorf("setting powerlimit for inverter %d did not succeed: %s", ctrlResp.ID, ctrlResp.Error)
	}
	return nil
}

package mystrom

import (
	"fmt"
	"net/http"
)

const turnOffRoute = "/relay?state=0"

type MyStrom struct {
	Endpoint string
	Client   http.Client
}

// return a new mystrom controller, or nil, if endpoint is set to ""
func NewMystrom(endpoint string) *MyStrom {
	if endpoint == "" {
		return nil
	}
	ms := MyStrom{
		Endpoint: endpoint,
		Client:   *http.DefaultClient,
	}
	return &ms
}

// Disable the mystrom plug
func (ms *MyStrom) Disable() error {
	req, err := http.NewRequest(http.MethodGet, ms.Endpoint+turnOffRoute, nil)
	if err != nil {
		return err
	}
	resp, err := ms.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("got status %d from mystrom, but expect 2XX", resp.StatusCode)
	}
	return nil
}

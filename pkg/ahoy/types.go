package ahoy

type InverterInfo struct {
	ID             int           `json:"id"`
	Enabled        bool          `json:"enabled"`
	Name           string        `json:"name"`
	Serial         string        `json:"serial"`
	Version        string        `json:"version"`
	PowerLimitRead int           `json:"power_limit_read"`
	PowerLimitAck  bool          `json:"power_limit_ack"`
	MaxPwr         int           `json:"max_pwr"`
	TsLastSuccess  int           `json:"ts_last_success"`
	Generation     int           `json:"generation"`
	Status         int           `json:"status"`
	AlarmCnt       int           `json:"alarm_cnt"`
	Rssi           int           `json:"rssi"`
	TsMaxAcPwr     int           `json:"ts_max_ac_pwr"`
	Ch             [][]int       `json:"ch"`
	ChName         []string      `json:"ch_name"`
	ChMaxPwr       []interface{} `json:"ch_max_pwr"`
}

type CtrlRequest struct {
	ID    int    `json:"id"`
	Token string `json:"token"`
	Cmd   string `json:"cmd"`
	Val   string `json:"val"`
}

type CtrlResponse struct {
	Success bool   `json:"success"`
	ID      int    `json:"id"`
	Error   string `json:"error"`
}

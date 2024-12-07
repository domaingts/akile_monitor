package model

type Data struct {
	Host      *Host      `json:"Host"`
	State     *HostState `json:"State"`
	Timestamp int64      `json:"TimeStamp"`
}

type HostState struct {
	CPU            float64 `json:"CPU"`
	MemUsed        uint64  `json:"MemUsed"`
	SwapUsed       uint64  `json:"SwapUsed"`
	NetInTransfer  uint64  `json:"NetInTransfer"`
	NetOutTransfer uint64  `json:"NetOutTransfer"`
	NetInSpeed     uint64  `json:"NetInSpeed"`
	NetOutSpeed    uint64  `json:"NetOutSpeed"`
	Uptime         uint64  `json:"Uptime"`
	Load1          float64 `json:"Load1"`
	Load5          float64 `json:"Load5"`
	Load15         float64 `json:"Load15"`
}

type Host struct {
	Name            string   `json:"Name"`
	Platform        string   `json:"Platform"`
	PlatformVersion string   `json:"PlatformVersion"`
	CPU             []string `json:"CPU"`
	MemTotal        uint64   `json:"MemTotal"`
	SwapTotal       uint64   `json:"SwapTotal"`
	Arch            string   `json:"Arch"`
	Virtualization  string   `json:"Virtualization"`
	BootTime        uint64   `json:"BootTime"`
}

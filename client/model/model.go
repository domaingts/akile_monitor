package model

type Data struct {
	Host      *Host      `json:"host"`
	State     *HostState `json:"state"`
	Timestamp int64      `json:"timestamp"`
}

type HostState struct {
	CPU            float64 `json:"cpu"`
	MemUsed        uint64  `json:"mem_used"`
	SwapUsed       uint64  `json:"swap_used"`
	NetInTransfer  uint64  `json:"net_in_transfer"`
	NetOutTransfer uint64  `json:"net_out_transfer"`
	NetInSpeed     uint64  `json:"net_in_speed"`
	NetOutSpeed    uint64  `json:"net_out_speed"`
	Uptime         uint64  `json:"uptime"`
	Load1          float64 `json:"load1"`
	Load5          float64 `json:"load5"`
	Load15         float64 `json:"load15"`
}

type Host struct {
	Name            string   `json:"name"`
	Platform        string   `json:"platform"`
	PlatformVersion string   `json:"platform_version"`
	CPU             []string `json:"cpu"`
	MemTotal        uint64   `json:"mem_total"`
	SwapTotal       uint64   `json:"swap_total"`
	Arch            string   `json:"arch"`
	Virtualization  string   `json:"virtualization"`
	BootTime        uint64   `json:"boot_time"`
}

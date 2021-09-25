package serverquery

type ApplicationStatus struct {
	Running bool `json:"Running"`
}

type ServerInfo struct {
	General           *GeneralInformation     `json:"General"`
	AppInfo           *ApplicationInformation `json:"Application"`
	ServiceInfo       *ServiceInformation     `json:"Service"`
	SystemInformation *SystemInformation      `json:"System"`
}

type GeneralInformation struct {
	Application string `json:"Application"`
	LastUpdated string `json:"Last-Updated"`
	Created     string `json:"Created"`
	Name        string `json:"Name"`
	ID          string `json:"ID"`
	OwnerID     string `json:"Owner-ID"`
	OwnerName   string `json:"Owner-Name"`
	ClientPort  int    `json:"Client-Port"`
	QueryPort   int    `json:"Query-Port,omitempty"`
	QueryType   string `json:"Query-Type,omityempty"`
	IP          string `json:"IP"`
	HostOS      string `json:"Host-OS"`
}

type ApplicationInformation struct {
	CurrentPlayers    int       `json:"Current-Players"`
	MaxPlayers        int       `json:"Max-Players"`
	Map               string    `json:"Map,omitempty"`
	PasswordProtected bool      `json:"Password-Protected"`
	VAC               bool      `json:"VAC,omitempty"`
	Players           []*Player `json:"Player-Info,omitempty"`
}

type Player struct {
	Name     string  `json:"Name"`
	Duration float32 `json:"Duration,omitempty"`
}

type ServiceInformation struct {
	Provider     string `json:"Provider"`
	HardwareType string `json:"Hardware-Type"`
	Region       string `json:"Region"`
}

type SystemInformation struct {
	Uptime string             `json:"Uptime"`
	CPU    []*CPUInformation  `json:"CPUs"`
	Memory *MemoryInformation `json:"Memory"`
	Disk   *DiskInformation   `json:"Disk"`
}

type MemoryInformation struct {
	Total       uint64  `json:"Total"`
	Available   uint64  `json:"Available"`
	Used        uint64  `json:"Used"`
	UsedPercent float64 `json:"Percent-Used"`
	Free        uint64  `json:"Free"`
}

type DiskInformation struct {
	Total       uint64  `json:"total"`
	Free        uint64  `json:"Free"`
	Used        uint64  `json:"Used"`
	UsedPercent float64 `json:"Percent-Used"`
}

type CPUInformation struct {
	CPU       int32   `json:"Cpu"`
	Vendor    string  `json:"Vendor"`
	Family    string  `json:"Family"`
	Cores     int32   `json:"Core(s)"`
	ModelName string  `json:"Model-Name"`
	Mhz       float64 `json:"Mhz"`
	CacheSize int32   `json:"Cache-Size"`
}

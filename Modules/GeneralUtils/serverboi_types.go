package generalutils

type BaseServer struct {
	ServerID    string `json:"ServerID"`
	Application string `json:"Application"`
	ServerName  string `json:"ServerName"`
	Service     string `json:"Service"`
	Owner       string `json:"Owner"`
	OwnerID     string `json:"OwnerID"`
	Port        int    `json:"Port"`
}

type ServerBoiRegion struct {
	Emoji   string
	Name    string
	Service string
	// Name of region in cloud provider
	ServiceName string
	Geolocation string
}

type AWSServer struct {
	ServerID     string `json:"ServerID"`
	Application  string `json:"Application"`
	ServerName   string `json:"ServerName"`
	Owner        string `json:"Owner"`
	OwnerID      string `json:"OwnerID"`
	Service      string `json:"Service"`
	AWSAccountID string `json:"AWSAccountID"`
	InstanceID   string `json:"InstanceID"`
	InstanceType string `json:"InstanceType"`
	Region       string `json:"Region"`
	Port         int    `json:"Port"`
}

// type LinodeServer struct {
// 	ServiceInfo LinodeService
// }

type LinodeServer struct {
	ServerID    string `json:"ServerID"`
	Application string `json:"Application"`
	ServerName  string `json:"ServerName"`
	Owner       string `json:"Owner"`
	OwnerID     string `json:"OwnerID"`
	Service     string `json:"Service"`
	Port        int    `json:"Port"`
	LinodeID    int    `json:"LinodeID"`
	ApiKey      string `json:"ApiKey"`
	LinodeType  string `json:"LinodeType"`
	Location    string `json:"Location"`
}

type ServerDescribeResponse struct {
	ServerType string
	State      string
	IP         string
}

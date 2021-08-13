package main

type ServerBoiServer struct {
	ServerID string            `json:"ServerID"`
	Game     string            `json:"Game"`
	Name     string            `json:"Name"`
	Owner    string            `json:"Owner"`
	OwnerID  string            `json:"OwnerID"`
	Service  map[string]string `json:"Service"`
	Port     int               `json:"Port"`
}

type ServerBoiRegion struct {
	Emoji   string
	Name    string
	Service string
	// Name of region in cloud provider
	ServiceName string
	Geolocation string
}

type AWSService struct {
	Name       string `json:"Name"`
	AccountID  string `json:"AccountID"`
	Region     string `json:"Region"`
	InstanceID string `json:"InstanceID"`
}

type AWSServer struct {
	ServiceInfo AWSService
}

type LinodeServer struct {
	ServiceInfo LinodeService
}

type LinodeService struct {
	Name       string
	LinodeID   int
	ApiKey     string
	LinodeType string
	Location   string
}

type ServerDescribeResponse struct {
	ServerType string
	State      string
	IP         string
}

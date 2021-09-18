package responseutils

import (
	"log"
	"strings"
)

var (
	awsRegionToSBRegion = map[string]RegionInfo{
		"us-west-2": {
			Name:     "US-West",
			Location: "Oregon",
			Emoji:    "🇺🇸",
		},
		"us-west-1": {
			Name:     "US-West",
			Location: "N. California",
			Emoji:    "🇺🇸",
		},
		"us-east-2": {
			Name:     "US-East",
			Location: "Ohio",
			Emoji:    "🇺🇸",
		},
		"us-east-1": {
			Name:     "US-East",
			Location: "Virginia",
			Emoji:    "🇺🇸",
		},
	}
	linodeLocationToSBRegion = map[string]RegionInfo{
		"us-west": {
			Location: "N. California",
			Name:     "US-West",
			Emoji:    "🇺🇸",
		},
		"us-east": {
			Location: "New York",
			Name:     "US-East",
			Emoji:    "🇺🇸",
		},
		"us-southeast": {
			Location: "Georgia",
			Name:     "US-South",
			Emoji:    "🇺🇸",
		},
		"us-central": {
			Location: "Texas",
			Name:     "US-Central",
			Emoji:    "🇺🇸",
		},
	}
)

type RegionInfo struct {
	Name     string
	Location string
	Emoji    string
}

func FormRegionInfo(service string, serviceRegion string) RegionInfo {
	switch strings.ToLower(service) {
	case "aws":
		return awsRegionToSBRegion[serviceRegion]
	case "linode":
		return linodeLocationToSBRegion[serviceRegion]
	default:
		log.Fatalf("No entry for region")
		return RegionInfo{}
	}
}

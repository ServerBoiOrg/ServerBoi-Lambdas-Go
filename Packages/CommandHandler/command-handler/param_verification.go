package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	gu "generalutils"
)

// Function to check if given service is supported for Serverboi
func verifyService(s string) error {
	service := strings.ToLower(s)

	switch service {
	case "aws":
		return nil
	case "linode":
		return nil
	default:
		return errors.New(fmt.Sprintf("* service: Unknown service `%v`", s))
	}
}

// Verifies region is a valid region for the service
func verifyRegion(s string, r string) error {
	service := strings.ToLower(s)
	region := strings.ToLower(r)

	switch service {
	case "aws":
		return verifyAWSRegion(region)
	case "linode":
		return nil
	default:
		return errors.New("* region: Valid service is required to check region")
	}
}

// Verifies the provided region is either an actual AWS region or a Serverboi Logical regions
func verifyAWSRegion(region string) error {
	log.Printf("Checking region %v", region)

	for _, awsRegion := range gu.AWSRegions {
		if region == awsRegion {
			return nil
		}
	}

	for _, serverboiRegion := range gu.AWSRegions {
		if region == serverboiRegion {
			return nil
		}
	}

	log.Printf("Given regions is not valid")
	return errors.New(fmt.Sprintf("* region: `%v` is not an AWS Region or ServerBoi Region", region))
}

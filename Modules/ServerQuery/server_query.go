package serverquery

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func ServerDataQuery(ip string) (*ServerInfo, error) {
	data := ServerInfo{}
	address := fmt.Sprintf("http://%v:7032/info", ip)
	resp, err := http.Get(address)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			json.NewDecoder(resp.Body).Decode(&data)
			return &data, nil
		} else {
			err = errors.New("Bad response")
		}
	}

	return &ServerInfo{}, err
}

func GetApplicationStatus(ip string) (*ApplicationStatus, error) {
	status := ApplicationStatus{}
	address := fmt.Sprintf("http://%v:7032/info", ip)
	resp, err := http.Get(address)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			json.NewDecoder(resp.Body).Decode(&status)
			return &status, nil
		} else {
			err = errors.New("Bad response")
		}
	}
	return &ApplicationStatus{}, err
}

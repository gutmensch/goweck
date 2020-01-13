package raumserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gutmensch/goweck/app"
)

type rendererState struct {
	Action string `json:"action,omitempty`
	Data   []struct {
		URI    string `json:"AVTransportURI,omitempty"`
		State  string `json:"TransportState,omitempty"`
		Status string `json:"TransportStatus,omitempty"`
	} `json:"data,omitempty"`
	Error   bool   `json:"error,omitempty`
	Message string `json:"msg,omitempty"`
}

// this is madnessting for sure
type zoneData struct {
	Action string `json:"action,omitempty`
	Data   struct {
		ZoneConfig struct {
			Zones []struct {
				Zone []struct {
					Self struct {
						Udn string `json:"udn,omitempty"`
					} `json:"$,omitempty"`
					Room []struct {
						Self struct {
							Name       string `json:"name,omitempty"`
							PowerState string `json:"powerState,omitempty"`
							Udn        string `json:"udn,omitempty"`
						} `json:"$,omitempty"`
					} `json:"room,omitempty"`
				} `json:"zone,omitempty"`
			} `json:"zones,omitempty"`
		} `json:"zoneConfig,omitempty"`
	} `json:"data,omitempty"`
	Error   bool   `json:"error,omitempty`
	Message string `json:"msg,omitempty"`
}

type simpleZone struct {
	Udn  string `json:"udn,omitempty"`
	Name string `json:"name,omitempty"`
}

var (
	RaumserverDebug, _ = strconv.ParseBool(app.GetEnvVar("DEBUG", "false"))
	RaumserverURI      = app.GetEnvVar("RAUMSERVER_URI", "http://localhost:3500/raumserver")
	NetClient          = &http.Client{Timeout: time.Second * 10}
)

func raumfeldAction(raumfeld string, uri string, params map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", raumfeld, uri), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	if RaumserverDebug {
		fmt.Println(req.URL.String())
	}
	resp, err := NetClient.Do(req)
	if RaumserverDebug && err != nil {
		fmt.Println(resp)
	}
	if resp != nil {
		respData, err := ioutil.ReadAll(resp.Body)
		return respData, err
	}
	return nil, errors.New("request to raumserver failed")
}

func PlayRaumfeldStream(zoneUUID string, streamURL string) error {
	params := map[string]string{
		"id":    zoneUUID,
		"value": streamURL,
	}
	_, err := raumfeldAction(RaumserverURI, "/controller/loadUri", params)
	return err
}

func PlayRaumfeldPlaylist(zoneUUID string, playlistName string) error {
	params := map[string]string{
		"id":    zoneUUID,
		"value": playlistName,
	}
	_, err := raumfeldAction(RaumserverURI, "/controller/loadPlaylist", params)
	return err
}

func PlayRaumfeldContainer(zoneUUID string, containerName string) error {
	params := map[string]string{
		"id":    zoneUUID,
		"value": containerName,
	}
	_, err := raumfeldAction(RaumserverURI, "/controller/loadPlaylist", params)
	return err
}

func AdjustRaumfeldVolume(zoneUUID string, volume int) error {
	params := map[string]string{
		"id":    zoneUUID,
		"scope": "zone",
		"value": fmt.Sprint(volume),
	}
	_, err := raumfeldAction(RaumserverURI, "/controller/setVolume", params)
	return err
}

func StopRaumfeldStream(zoneUUID string) error {
	params := map[string]string{
		"id": zoneUUID,
	}
	_, err := raumfeldAction(RaumserverURI, "/controller/stop", params)
	return err
}

func LeaveStandby(zoneUUID string) error {
	params := map[string]string{
		"id": zoneUUID,
	}
	_, err := raumfeldAction(RaumserverURI, "/controller/leaveStandby", params)
	return err
}

func EnterStandby(zoneUUID string) error {
	params := map[string]string{
		"id": zoneUUID,
	}
	_, err := raumfeldAction(RaumserverURI, "/controller/enterManualStandby", params)
	return err
}

func CheckTransportState(zoneUUID string) (bool, error) {
	params := map[string]string{
		"id": zoneUUID,
	}
	var rendererState rendererState
	data, err := raumfeldAction(RaumserverURI, "/data/getRendererState", params)
	err = json.Unmarshal(data, &rendererState)
	if RaumserverDebug {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(rendererState)
	}
	if len(rendererState.Data) == 0 {
		return false, err
	}
	if rendererState.Data[0].State != "PLAYING" {
		return false, err
	}
	return true, err
}

func GetZones() (zoneData, error) {
	params := map[string]string{}
	var zoneData zoneData
	data, err := raumfeldAction(RaumserverURI, "/data/getZoneConfig", params)
	if err != nil && RaumserverDebug {
		app.Log(err)
		return zoneData, err
	}
	err = json.Unmarshal(data, &zoneData)
	if RaumserverDebug {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(zoneData)
	}
	return zoneData, err
}

func GetZoneListJSON() (string, error) {
	var zones zoneData
	var nz []simpleZone
	var z simpleZone
	var friendlyName []string
	zones, err := GetZones()
	if err != nil {
		return "", err
	}
	for _, zone := range zones.Data.ZoneConfig.Zones {
		for _, realZone := range zone.Zone {
			friendlyName = []string{}
			z.Udn = realZone.Self.Udn
			for _, room := range realZone.Room {
				friendlyName = append(friendlyName, room.Self.Name)
			}
			z.Name = strings.Join(friendlyName, ", ")
			nz = append(nz, z)
		}
	}
	res, err := json.Marshal(nz)
	return string(res), err
}

func GetZoneName(uuid string) string {
	zoneName := "unknown"
	var friendlyName []string
	zoneData, err := GetZones()
	if err != nil {
		return zoneName
	}
	for _, zones := range zoneData.Data.ZoneConfig.Zones {
		for _, realZone := range zones.Zone {
			friendlyName = []string{}
			for _, room := range realZone.Room {
				friendlyName = append(friendlyName, room.Self.Name)
			}
			if realZone.Self.Udn == uuid {
				zoneName = strings.Join(friendlyName, ", ")
			}
		}
	}
	return zoneName
}

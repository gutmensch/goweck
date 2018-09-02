package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	"github.com/gregdel/pushover"
	"github.com/imdario/mergo"

	. "github.com/gutmensch/goweck/appbase"
)

type Alarm struct {
	ID            bson.ObjectId `bson:"_id,omitempty"           json:"_id,omitempty"`
	Enable        string        `bson:"enable,omitempty"        json:"enable,omitempty"`
	HourMinute    string        `bson:"hourMinute,omitempty"    json:"hourMinute,omitempty"`
	LastModified  string        `bson:"lastModified,omitempty"  json:"lastModified,omitempty"`
	WeekDays      string        `bson:"weekDays,omitempty"      json:"weekDays,omitempty"`
	WeekEnds      string        `bson:"weekEnds,omitempty"      json:"weekEnds,omitempty"`
	RaumserverUri string        `bson:"raumserverUri,omitempty" json:"raumserverUri,omitempty"`
	ZoneUuid      string        `bson:"zoneUuid,omitempty"      json:"zoneUuid,omitempty"`
	ZoneName      string        `bson:"zoneName,omitempty"      json:"zoneName,omitempty"`
	RadioChannel  string        `bson:"radioChannel,omitempty"  json:"radioChannel,omitempty"`
	VolumeStart   int           `bson:"volumeStart,omitempty"   json:"volumeStart,omitempty"`
	VolumeEnd     int           `bson:"volumeEnd,omitempty"     json:"volumeEnd,omitempty"`
	VolumeIncStep int           `bson:"volumeIncStep,omitempty" json:"volumeIncStep,omitempty"`
	VolumeIncInt  int           `bson:"volumeIncInt,omitempty"  json:"volumeIncInt,omitempty"`
	Timeout       int           `bson:"timeout,omitempty"       json:"timeout,omitempty"`
}

type RendererState struct {
	Action string `json:"action,omitempty`
	Data   []struct {
		Uri    string `json:"AVTransportURI,omitempty"`
		State  string `json:"TransportState,omitempty"`
		Status string `json:"TransportStatus,omitempty"`
	} `json:"data,omitempty"`
	Error   bool   `json:"error,omitempty`
	Message string `json:"msg,omitempty"`
}

// this is madnessting for sure
type ZoneData struct {
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

type SimpleZone struct {
	Udn  string `json:"udn,omitempty"`
	Name string `json:"name,omitempty"`
}

var (
	Debug, _           = strconv.ParseBool(GetEnvVar("DEBUG", "false"))
	RaumserverDebug, _ = strconv.ParseBool(GetEnvVar("RAUMSERVER_DEBUG", "false"))
	// DB parameters, needed for running
	MongoDbDrop, _ = strconv.ParseBool(GetEnvVar("MONGODB_DROP", "false"))
	MongoDb        = GetEnvVar("MONGODB_DATABASE", "goweck")
	MongoUri       = GetEnvVar("MONGODB_URI", "mongodb://127.0.0.1:27017")
	Listen         = GetEnvVar("LISTEN", ":8080")
	CheckInterval  = 5
	Database       *mgo.Database
	// defaults for raumserver connection. should be overridden by every specific alarm
	RaumserverUri = GetEnvVar("RAUMSERVER_URI", "http://qnap:3535/raumserver")
	RadioChannel  = GetEnvVar("RADIO_CHANNEL", "http://mp3channels.webradio.rockantenne.de/alternative")
	TimeZone      = GetEnvVar("TZ", "UTC")
	PushOverUser  = GetEnvVar("PUSHOVER_USER_TOKEN", "undefined")
	PushOverApp   = GetEnvVar("PUSHOVER_APP_TOKEN", "undefined")
	AlarmActive   = false
	NetClient     = &http.Client{Timeout: time.Second * 10}
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
	if RaumserverDebug {
		fmt.Println(resp)
	}
	respData, _ := ioutil.ReadAll(resp.Body)
	return respData, err
}

func playRaumfeldStream(alarm *Alarm) error {
	params := map[string]string{
		"id":    alarm.ZoneUuid,
		"value": alarm.RadioChannel,
	}
	_, err := raumfeldAction(alarm.RaumserverUri, "/controller/loadUri", params)
	return err
}

func playRaumfeldPlaylist(alarm *Alarm) error {
	params := map[string]string{
		"id":    alarm.ZoneUuid,
		"value": alarm.RadioChannel,
	}
	_, err := raumfeldAction(alarm.RaumserverUri, "/controller/loadPlaylist", params)
	return err
}

func adjustRaumfeldVolume(alarm *Alarm, volume int) error {
	params := map[string]string{
		"id":    alarm.ZoneUuid,
		"scope": "zone",
		"value": fmt.Sprint(volume),
	}
	_, err := raumfeldAction(alarm.RaumserverUri, "/controller/setVolume", params)
	return err
}

func stopRaumfeldStream(alarm *Alarm) error {
	params := map[string]string{
		"id": alarm.ZoneUuid,
	}
	_, err := raumfeldAction(alarm.RaumserverUri, "/controller/stop", params)
	return err
}

func leaveStandby(alarm *Alarm) error {
	params := map[string]string{
		"id": alarm.ZoneUuid,
	}
	_, err := raumfeldAction(alarm.RaumserverUri, "/controller/leaveStandby", params)
	return err
}

func enterStandby(alarm *Alarm) error {
	params := map[string]string{
		"id": alarm.ZoneUuid,
	}
	_, err := raumfeldAction(alarm.RaumserverUri, "/controller/enterManualStandby", params)
	return err
}

func checkTransportState(alarm *Alarm) (bool, error) {
	params := map[string]string{
		"id": alarm.ZoneUuid,
	}
	var rendererState RendererState
	data, err := raumfeldAction(alarm.RaumserverUri, "/data/getRendererState", params)
	err = json.Unmarshal(data, &rendererState)
	if Debug {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(rendererState)
	}
	if rendererState.Data[0].State != "PLAYING" {
		return false, err
	}
	return true, err
}

func getZones() (ZoneData, error) {
	params := map[string]string{}
	var zoneData ZoneData
	data, err := raumfeldAction(RaumserverUri, "/data/getZoneConfig", params)
	err = json.Unmarshal(data, &zoneData)
	if Debug {
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(zoneData)
	}
	return zoneData, err
}

func getZoneName(uuid string) string {
	zoneName := "unknown"
	var friendlyName []string
	zoneData, _ := getZones()
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

func isStream(s string) bool {
	validStream := regexp.MustCompile(`^https?://`)
	return validStream.MatchString(s)
}

func sendFallbackMessage() error {
	app := pushover.New(PushOverApp)
	recipient := pushover.NewRecipient(PushOverUser)
	message := pushover.NewMessageWithTitle("Wake up please!", "GoWeck")
	response, err := app.SendMessage(message, recipient)
	if Debug {
		fmt.Println(response)
	}
	return err
}

func executeAlarm(alarm *Alarm) {
	// play stream for wake up
	err := leaveStandby(alarm)
	Log(err)
	time.Sleep(10 * time.Second)
	err = adjustRaumfeldVolume(alarm, alarm.VolumeStart)
	Log(err)
	if isStream(alarm.RadioChannel) {
		err = playRaumfeldStream(alarm)
	} else {
		err = playRaumfeldPlaylist(alarm)
	}
	Log(err)
	// inc volume over time
	steps := alarm.VolumeEnd - alarm.VolumeStart
	err_count := 0
	for i := 1; i <= steps; i++ {
		running, _ := checkTransportState(alarm)
		if (i%5) == 0 && !running {
			err_count += 1
			if isStream(alarm.RadioChannel) {
				err = playRaumfeldStream(alarm)
			} else {
				err = playRaumfeldPlaylist(alarm)
			}
			Log(err)
		}
		if err_count >= 3 {
			// fallback: send pushover notification to wake up the guy
			err = sendFallbackMessage()
			Log(err)
			err_count = 0
		}
		if Debug {
			fmt.Println("adjusting volume to", alarm.VolumeStart+i)
		}
		err = adjustRaumfeldVolume(alarm, alarm.VolumeStart+i)
		Log(err)
		time.Sleep(time.Duration(alarm.VolumeIncInt) * time.Second)
	}
	// sleep for the rest of the alarm active time
	time.Sleep(time.Duration(alarm.Timeout-steps*alarm.VolumeIncInt) * time.Second)
	// cleanup and unset alarm
	err = adjustRaumfeldVolume(alarm, alarm.VolumeStart)
	Log(err)
	err = stopRaumfeldStream(alarm)
	Log(err)
	err = enterStandby(alarm)
	Log(err)
	AlarmActive = false
}

func pollAlarm() {
	var result Alarm
	c := Database.C("alarm").With(Database.Session.Copy())

	for {
		<-time.After(time.Duration(CheckInterval) * time.Second)

		if AlarmActive == true {
			fmt.Println("[pollAlarm] alarm is currently active")
			continue
		}

		loc, _ := time.LoadLocation(TimeZone)
		t := time.Now().In(loc)

		search := bson.M{}
		switch int(t.Weekday()) {
		case 0, 6:
			search = bson.M{
				"enable":     "true",
				"hourMinute": fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute()),
				"weekEnds":   strconv.FormatBool(true),
			}
		case 1, 2, 3, 4, 5:
			search = bson.M{
				"enable":     "true",
				"hourMinute": fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute()),
				"weekDays":   strconv.FormatBool(true),
			}
		}

		err := c.Find(search).One(&result)
		if err == mgo.ErrNotFound || err != nil {
			if Debug {
				fmt.Println(err.Error())
			}
			continue
		}

		fmt.Println("[pollAlarm] alarm for current time and day found, executing.")
		AlarmActive = true
		go executeAlarm(&result)
	}
}

func main() {
	session, err := mgo.Dial(MongoUri)
	defer session.Close()
	Fatal(err)

	session.SetMode(mgo.Monotonic, true)

	if MongoDbDrop {
		err = session.DB(MongoDb).DropDatabase()
		Log(err)
	}
	Database = session.DB(MongoDb)

	ensureIndices()

	// execute or change alarms periodically
	go pollAlarm()

	// http endpoint for dealing with alarms
	router := httpRouter()
	log.Fatal(http.ListenAndServe(Listen, router))
}

func ensureIndices() {
	c := Database.C("alarm").With(Database.Session.Copy())

	index := mgo.Index{
		Key:        []string{"enable", "hourMinute", "weekDays", "weekEnds"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err := c.EnsureIndex(index)
	Fatal(err)
}

func httpRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/alarms", http.HandlerFunc(AlarmListHandler)).Methods("GET")
	r.HandleFunc("/zones", http.HandlerFunc(ZoneListHandler)).Methods("GET")
	r.HandleFunc("/alarm", http.HandlerFunc(AlarmCreateHandler)).Methods("POST")
	r.HandleFunc("/alarm/{id:[0-9a-f]{24}}", http.HandlerFunc(AlarmUpdateHandler)).Methods("POST")
	r.HandleFunc("/alarm/{id:[0-9a-f]{24}}", http.HandlerFunc(AlarmDeleteHandler)).Methods("DELETE")
	r.HandleFunc("/", http.HandlerFunc(IndexHandler)).Methods("GET")
	r.HandleFunc("/index.html", http.HandlerFunc(IndexHandler)).Methods("GET")
	return r
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(IndexHtml))
}

func AlarmListHandler(w http.ResponseWriter, r *http.Request) {
	var results []Alarm
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.Find(bson.M{}).All(&results)
	Fatal(err)
	e, err := json.Marshal(results)
	Log(err)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%q\n", string(e))
}

func ZoneListHandler(w http.ResponseWriter, r *http.Request) {
	var zones ZoneData
	var nz []SimpleZone
	var z SimpleZone
	var friendlyName []string
	zones, err := getZones()
	Log(err)
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
	e, err := json.Marshal(nz)
	Log(err)
	fmt.Fprintf(w, "%q\n", string(e))
}

func AlarmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	Fatal(err)
	var new, old Alarm
	err = json.Unmarshal(body, &new)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	Fatal(err)
	vars := mux.Vars(r)
	c := Database.C("alarm").With(Database.Session.Copy())
	err = c.FindId(bson.ObjectIdHex(vars["id"])).One(&old)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = mergo.Merge(&new, old)
	Log(err)
	err = c.UpdateId(bson.ObjectIdHex(vars["id"]), &new)
	if err != nil {
		Log(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func AlarmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.RemoveId(bson.ObjectIdHex(vars["id"]))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func AlarmCreateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	Fatal(err)
	var new, def, result Alarm
	err = json.Unmarshal(body, &new)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c := Database.C("alarm").With(Database.Session.Copy())
	dupAlarm := c.Find(bson.M{
		"hourMinute": new.HourMinute,
		"weekDays":   new.WeekDays,
		"weekEnds":   new.WeekEnds,
	}).One(&result)

	// bail out early if anything else then NotFound
	if dupAlarm != mgo.ErrNotFound {
		w.WriteHeader(http.StatusConflict)
		return
	}

	def = Alarm{
		Enable:        strconv.FormatBool(true),
		HourMinute:    "08:00",
		LastModified:  time.Now().String(),
		WeekDays:      strconv.FormatBool(true),
		WeekEnds:      strconv.FormatBool(false),
		RaumserverUri: RaumserverUri,
		RadioChannel:  RadioChannel,
		ZoneUuid:      "",
		ZoneName:      getZoneName(new.ZoneUuid),
		VolumeStart:   5,
		VolumeEnd:     45,
		VolumeIncStep: 1,
		VolumeIncInt:  20,
		Timeout:       7200,
	}
	err = mergo.Merge(&new, def)
	if err != nil {
		Log(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = c.Insert(&new)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

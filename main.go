package main

import (
	"encoding/json"
	"errors"
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
	"github.com/gutmensch/goweck/app"
	"github.com/imdario/mergo"
)

type stream struct {
	ID   bson.ObjectId `bson:"_id,omitempty"    json:"_id,omitempty"`
	Name string        `bson:"name,omitempty"   json:"name,omitempty"`
	URL  string        `bson:"url,omitempty"    json:"url,omitempty"`
}

type alarm struct {
	ID            bson.ObjectId `bson:"_id,omitempty"           json:"_id,omitempty"`
	Status        string        `bson:"status,omitempty"        json:"status,omitempty"`
	HourMinute    string        `bson:"hourMinute,omitempty"    json:"hourMinute,omitempty"`
	LastModified  string        `bson:"lastModified,omitempty"  json:"lastModified,omitempty"`
	WeekDays      string        `bson:"weekDays,omitempty"      json:"weekDays,omitempty"`
	WeekEnds      string        `bson:"weekEnds,omitempty"      json:"weekEnds,omitempty"`
	RaumserverURI string        `bson:"raumserverUri,omitempty" json:"raumserverUri,omitempty"`
	ZoneUUID      string        `bson:"zoneUuid,omitempty"      json:"zoneUuid,omitempty"`
	ZoneName      string        `bson:"zoneName,omitempty"      json:"zoneName,omitempty"`
	StreamName    string        `bson:"streamName,omitempty"    json:"streamName,omitempty"`
	VolumeStart   int           `bson:"volumeStart,omitempty"   json:"volumeStart,omitempty"`
	VolumeEnd     int           `bson:"volumeEnd,omitempty"     json:"volumeEnd,omitempty"`
	VolumeIncStep int           `bson:"volumeIncStep,omitempty" json:"volumeIncStep,omitempty"`
	VolumeIncInt  int           `bson:"volumeIncInt,omitempty"  json:"volumeIncInt,omitempty"`
	Timeout       int           `bson:"timeout,omitempty"       json:"timeout,omitempty"`
}

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
	Debug, _             = strconv.ParseBool(app.GetEnvVar("DEBUG", "false"))
	RaumserverDebug, _   = strconv.ParseBool(app.GetEnvVar("RAUMSERVER_DEBUG", "false"))
	MongoDbDrop, _       = strconv.ParseBool(app.GetEnvVar("MONGODB_DROP", "false"))
	MongoDb              = app.GetEnvVar("MONGODB_DATABASE", "goweck")
	MongoURI             = app.GetEnvVar("MONGODB_URI", "mongodb://127.0.0.1:27017")
	Listen               = app.GetEnvVar("LISTEN", ":8080")
	CheckInterval        = 5
	Database             *mgo.Database
	enableDeepStandby, _ = strconv.ParseBool(app.GetEnvVar("DEEP_STANDBY", "false"))
	RaumserverURI        = app.GetEnvVar("RAUMSERVER_URI", "http://qnap:3535/raumserver")
	TimeZone             = app.GetEnvVar("TZ", "UTC")
	PushOverUser         = app.GetEnvVar("PUSHOVER_USER_TOKEN", "undefined")
	PushOverApp          = app.GetEnvVar("PUSHOVER_APP_TOKEN", "undefined")
	AlarmActive          = false
	NetClient            = &http.Client{Timeout: time.Second * 10}
	RadioStreams         = []stream{
		stream{
			Name: "Rock Antenne",
			URL:  "http://mp3channels.webradio.rockantenne.de/alternative",
		},
		stream{
			Name: "Radio 21 Rock",
			URL:  "http://188.94.97.91/radio21.mp3",
		},
		stream{
			Name: "BRF 91.4",
			URL:  "http://stream.berliner-rundfunk.de/brf/mp3-128/internetradio",
		},
		stream{
			Name: "RadioEins",
			URL:  "http://radioeins.de/stream",
		},
		stream{
			Name: "Fritz",
			URL:  "http://fritz.de/livemp3",
		},
	}
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

func getStreamURL(name string) string {
	for _, s := range RadioStreams {
		if s.Name == name {
			return s.URL
		}
	}
	return "not found"
}

func playRaumfeldStream(alarm *alarm) error {
	params := map[string]string{
		"id":    alarm.ZoneUUID,
		"value": getStreamURL(alarm.StreamName),
	}
	_, err := raumfeldAction(alarm.RaumserverURI, "/controller/loadUri", params)
	return err
}

func playRaumfeldPlaylist(alarm *alarm) error {
	params := map[string]string{
		"id":    alarm.ZoneUUID,
		"value": alarm.StreamName,
	}
	_, err := raumfeldAction(alarm.RaumserverURI, "/controller/loadPlaylist", params)
	return err
}

func adjustRaumfeldVolume(alarm *alarm, volume int) error {
	params := map[string]string{
		"id":    alarm.ZoneUUID,
		"scope": "zone",
		"value": fmt.Sprint(volume),
	}
	_, err := raumfeldAction(alarm.RaumserverURI, "/controller/setVolume", params)
	return err
}

func stopRaumfeldStream(alarm *alarm) error {
	params := map[string]string{
		"id": alarm.ZoneUUID,
	}
	_, err := raumfeldAction(alarm.RaumserverURI, "/controller/stop", params)
	return err
}

func leaveStandby(alarm *alarm) error {
	params := map[string]string{
		"id": alarm.ZoneUUID,
	}
	_, err := raumfeldAction(alarm.RaumserverURI, "/controller/leaveStandby", params)
	return err
}

func enterStandby(alarm *alarm) error {
	params := map[string]string{
		"id": alarm.ZoneUUID,
	}
	_, err := raumfeldAction(alarm.RaumserverURI, "/controller/enterManualStandby", params)
	return err
}

func checkTransportState(alarm *alarm) (bool, error) {
	params := map[string]string{
		"id": alarm.ZoneUUID,
	}
	var rendererState rendererState
	data, err := raumfeldAction(alarm.RaumserverURI, "/data/getRendererState", params)
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

func getZones() (zoneData, error) {
	params := map[string]string{}
	var zoneData zoneData
	data, err := raumfeldAction(RaumserverURI, "/data/getZoneConfig", params)
	if err != nil && Debug {
		app.Log(err)
		return zoneData, err
	}
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
	zoneData, err := getZones()
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

func executeAlarm(alarm *alarm) {
	// play stream for wake up
	if enableDeepStandby {
		leaveStandby(alarm)
	}

	time.Sleep(10 * time.Second)
	err := adjustRaumfeldVolume(alarm, alarm.VolumeStart)
	app.Log(err)
	err = playRaumfeldStream(alarm)
	app.Log(err)
	// inc volume over time
	steps := alarm.VolumeEnd - alarm.VolumeStart
	errCount := 0
	for i := 1; i <= steps; i++ {
		running, _ := checkTransportState(alarm)
		if (i%5) == 0 && !running {
			errCount++
			err = playRaumfeldStream(alarm)
			app.Log(err)
		}
		if errCount >= 3 {
			// fallback: send pushover notification to wake up the guy
			err = sendFallbackMessage()
			app.Log(err)
			errCount = 0
		}
		if Debug {
			fmt.Println("adjusting volume to", alarm.VolumeStart+i)
		}
		err = adjustRaumfeldVolume(alarm, alarm.VolumeStart+i)
		app.Log(err)
		time.Sleep(time.Duration(alarm.VolumeIncInt) * time.Second)
	}
	// sleep for the rest of the alarm active time
	time.Sleep(time.Duration(alarm.Timeout-steps*alarm.VolumeIncInt) * time.Second)
	// cleanup and unset alarm
	err = adjustRaumfeldVolume(alarm, alarm.VolumeStart)
	app.Log(err)
	err = stopRaumfeldStream(alarm)
	app.Log(err)
	if enableDeepStandby {
		enterStandby(alarm)
	}
	AlarmActive = false
}

func pollAlarm() {
	var result alarm
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
				"status":     "active",
				"hourMinute": fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute()),
				"weekEnds":   strconv.FormatBool(true),
			}
		case 1, 2, 3, 4, 5:
			search = bson.M{
				"status":     "active",
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
	session, err := mgo.Dial(MongoURI)
	defer session.Close()
	app.Fatal(err)

	session.SetMode(mgo.Monotonic, true)

	if MongoDbDrop {
		err = session.DB(MongoDb).DropDatabase()
		app.Log(err)
	}
	Database = session.DB(MongoDb)

	ensureIndices()

	populateStreams()

	// execute or change alarms periodically
	go pollAlarm()

	// http endpoint for dealing with alarms
	router := httpRouter()
	log.Fatal(http.ListenAndServe(Listen, router))
}

func ensureIndices() {
	c := Database.C("alarm").With(Database.Session.Copy())
	index := mgo.Index{
		Key:        []string{"status", "hourMinute", "weekDays", "weekEnds"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	app.Fatal(err)

	c = Database.C("stream").With(Database.Session.Copy())
	index = mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(index)
	app.Fatal(err)
}

func populateStreams() {
	var result stream
	for _, stream := range RadioStreams {
		c := Database.C("stream").With(Database.Session.Copy())
		dupStream := c.Find(bson.M{
			"name": stream.Name,
		}).One(&result)
		if dupStream != mgo.ErrNotFound {
			continue
		} else {
			err := c.Insert(stream)
			if err != nil {
				app.Log(err)
			}
		}
	}
}

func httpRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/alarm/all", http.HandlerFunc(alarmListHandler)).Methods("GET")
	r.HandleFunc("/zone/all", http.HandlerFunc(zoneListHandler)).Methods("GET")
	r.HandleFunc("/stream/all", http.HandlerFunc(streamListHandler)).Methods("GET")
	r.HandleFunc("/alarm/create", http.HandlerFunc(alarmCreateHandler)).Methods("POST")
	r.HandleFunc("/alarm/stop", http.HandlerFunc(alarmStopHandler)).Methods("POST")
	r.HandleFunc("/alarm/update/{id:[0-9a-f]{24}}", http.HandlerFunc(alarmUpdateHandler)).Methods("POST")
	r.HandleFunc("/alarm/delete/{id:[0-9a-f]{24}}", http.HandlerFunc(alarmDeleteHandler)).Methods("DELETE")
	r.HandleFunc("/", http.HandlerFunc(indexHandler)).Methods("GET")
	r.HandleFunc("/index.html", http.HandlerFunc(indexHandler)).Methods("GET")
	return r
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(indexHTML))
}

func alarmListHandler(w http.ResponseWriter, r *http.Request) {
	var results []alarm
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.Find(bson.M{}).All(&results)
	app.Fatal(err)
	e, err := json.Marshal(results)
	app.Log(err)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%q\n", string(e))
}

func zoneListHandler(w http.ResponseWriter, r *http.Request) {
	var zones zoneData
	var nz []simpleZone
	var z simpleZone
	var friendlyName []string
	zones, err := getZones()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
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
	e, err := json.Marshal(nz)
	app.Log(err)
	fmt.Fprintf(w, "%q\n", string(e))
}

func streamListHandler(w http.ResponseWriter, r *http.Request) {
	var results []stream
	c := Database.C("stream").With(Database.Session.Copy())
	err := c.Find(bson.M{}).All(&results)
	app.Fatal(err)
	e, err := json.Marshal(results)
	app.Log(err)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%q\n", string(e))
}

func alarmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	app.Fatal(err)
	var new, old alarm
	err = json.Unmarshal(body, &new)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	app.Fatal(err)
	vars := mux.Vars(r)
	c := Database.C("alarm").With(Database.Session.Copy())
	err = c.FindId(bson.ObjectIdHex(vars["id"])).One(&old)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = mergo.Merge(&new, old)
	app.Log(err)
	err = c.UpdateId(bson.ObjectIdHex(vars["id"]), &new)
	if err != nil {
		app.Log(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func alarmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.RemoveId(bson.ObjectIdHex(vars["id"]))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func alarmStopHandler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
}

func alarmCreateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	app.Fatal(err)
	var new, def, result alarm
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

	def = alarm{
		Status:        "Active",
		HourMinute:    "08:00",
		LastModified:  time.Now().String(),
		WeekDays:      strconv.FormatBool(true),
		WeekEnds:      strconv.FormatBool(false),
		RaumserverURI: RaumserverURI,
		StreamName:    "NOT_SET",
		ZoneUUID:      "",
		ZoneName:      getZoneName(new.ZoneUUID),
		VolumeStart:   5,
		VolumeEnd:     40,
		VolumeIncStep: 1,
		VolumeIncInt:  20,
		Timeout:       7200,
	}
	err = mergo.Merge(&new, def)
	if err != nil {
		app.Log(err)
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

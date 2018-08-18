package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	//"os"
	//"errors"

	"time"
	//"strconv"

	. "github.com/gutmensch/goweck/appbase"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

type Alarm struct {
	ID            bson.ObjectId `bson:"_id,omitempty" json:"_id,omitempty"`
	Enable        bool          `bson:"enable"        json:"enable"`
	HourMinute    string        `bson:"hourMinute"    json:"hourMinute"`
	CreatedAt     time.Time     `bson:"createdAt"     json:"createdAt"`
	WeekDays      bool          `bson:"weekDays"      json:"weekDays"`
	WeekEnds      bool          `bson:"weekEnds"      json:"weekEnds"`
	RaumserverUri string        `bson:"raumserverUri" json:"raumserverUri"`
	AlarmZone     string        `bson:"alarmZone"     json:"alarmZone"`
	RadioChannel  string        `bson:"radioChannel"  json:"radioChannel"`
	VolumeStart   int           `bson:"volumeStart"   json:"volumeStart"`
	VolumeEnd     int           `bson:"volumeEnd"     json:"volumeEnd"`
	VolumeIncStep int           `bson:"volumeIncStep" json:"volumeIncStep"`
	VolumeIncInt  int           `bson:"volumeIncInt"  json:"volumeIncInt"`
	Timeout       int           `bson:"timeout"       json:"timeout"`
}

var (
	Debug, _           = strconv.ParseBool(GetEnvVar("DEBUG", "false"))
	RaumserverDebug, _ = strconv.ParseBool(GetEnvVar("RAUMSERVER_DEBUG", "false"))
	MongoDbDrop, _     = strconv.ParseBool(GetEnvVar("MONGODB_DROP", "false"))
	MongoDb            = GetEnvVar("MONGODB_DATABASE", "goweck")
	MongoUri           = GetEnvVar("MONGODB_URI", "mongodb://127.0.0.1:27017")
	Listen             = GetEnvVar("LISTEN", ":8080")
	CheckInterval      = 5
	Database           *mgo.Database
	RaumserverUri      = GetEnvVar("RAUMSERVER_URI", "http://127.0.0.1:3535/raumserver")
	AlarmActive        = false
	NetClient          = &http.Client{Timeout: time.Second * 10}
)

func raumfeldAction(raumfeld string, uri string, params map[string]string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", raumfeld, uri), nil)
	if err != nil {
		return err
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

	return err
}

func playRaumfeldStream(alarm *Alarm) error {
	params := map[string]string{
		"id":    alarm.AlarmZone,
		"value": alarm.RadioChannel,
	}
	err := raumfeldAction(alarm.RaumserverUri, "/controller/loadUri", params)
	return err
}

func adjustRaumfeldVolume(alarm *Alarm, volume int) error {
	params := map[string]string{
		"id":    alarm.AlarmZone,
		"scope": "zone",
		"value": fmt.Sprint(volume),
	}
	err := raumfeldAction(alarm.RaumserverUri, "/controller/setVolume", params)
	return err
}

func stopRaumfeldStream(alarm *Alarm) error {
	params := map[string]string{
		"id": alarm.AlarmZone,
	}
	err := raumfeldAction(alarm.RaumserverUri, "/controller/stop", params)
	return err
}

func executeAlarm(alarm *Alarm) {
	// play stream for wake up
	err := adjustRaumfeldVolume(alarm, alarm.VolumeStart)
	Log(err)
	err = playRaumfeldStream(alarm)
	Log(err)
	// inc volume over time
	steps := alarm.VolumeEnd - alarm.VolumeStart
	for i := 1; i <= steps; i++ {
		fmt.Println("adjusting volume to", alarm.VolumeStart+i)
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

		t := time.Now()

		weekDays, weekEnds := true, true
		switch int(t.Weekday()) {
		case 0, 6:
			weekDays, weekEnds = false, true
		case 1, 2, 3, 4, 5:
			weekDays, weekEnds = true, false
		}

		err := c.Find(bson.M{
			"hourMinute": fmt.Sprintf("%02d%02d", t.Hour(), t.Minute()),
			"weekDays":   weekDays,
			"weekEnds":   weekEnds,
		}).One(&result)
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

	saveAlarm("1601", true, false, "uuid:C43C1A1D-AED1-472B-B0D0-210B7925000E", "http://mp3channels.webradio.rockantenne.de/alternative")

	// execute or change alarms periodically
	go pollAlarm()

	// http endpoint for dealing with alarms
	mux := httpRouter()
	log.Fatal(http.ListenAndServe(Listen, mux))
}

func ensureIndices() {
	c := Database.C("alarm").With(Database.Session.Copy())

	index := mgo.Index{
		Key:        []string{"hourMinute", "weekDays", "weekEnds"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err := c.EnsureIndex(index)
	Fatal(err)
}

func saveAlarm(hourMinute string, weekDays bool, weekEnds bool, zoneId string, radioChannel string) {
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.Insert(&Alarm{
		Enable:        true,
		HourMinute:    hourMinute,
		CreatedAt:     time.Now(),
		WeekDays:      weekDays,
		WeekEnds:      weekEnds,
		RaumserverUri: RaumserverUri,
		RadioChannel:  radioChannel,
		AlarmZone:     zoneId,
		VolumeStart:   15,
		VolumeEnd:     45,
		VolumeIncStep: 1,
		VolumeIncInt:  20,
		Timeout:       3600,
	})
	Fatal(err)
}

func httpRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/alarm", http.HandlerFunc(AlarmGetHandler)).Methods("GET")
	staticFileDirectory := http.Dir("./assets/")
	staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/assets/").Handler(staticFileHandler).Methods("GET")

	//r.Get("/alarm", http.HandlerFunc(AlarmGetHandler))
	//r.Post("/alarm", http.HandlerFunc(AlarmPostHandler))
	//r.Handle("/", http.HandlerFunc(DocumentationHandler))

	//staticFileDirectory := http.Dir("./assets/")
	//staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFileDirectory))
	//r.Prefix("/assets/").Handl

	return r
}

func AlarmGetHandler(w http.ResponseWriter, r *http.Request) {
	var results []Alarm
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.Find(bson.M{}).All(&results)
	Fatal(err)
	e, err := json.Marshal(results)
	Log(err)
	fmt.Fprintf(w, "%q\n", string(e))
}

func AlarmPostHandler(w http.ResponseWriter, r *http.Request) {
}

func DocumentationHandler(w http.ResponseWriter, r *http.Request) {
}

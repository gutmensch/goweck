package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	//"os"
	//"errors"
	"strings"
	"time"
	//"strconv"

	. "github.com/gutmensch/goweck/appbase"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/go-zoo/bone"
)

type Alarm struct {
	ID            bson.ObjectId `bson:"_id,omitempty"`
	HourMinute    string
	CreatedAt     time.Time
	WeekDays      bool
	WeekEnds      bool
	RaumserverUrl string
	AlarmZone     string
	RadioChannel  string
	VolumeStart   int
	VolumeEnd     int
	VolumeIncInt  int
	Timeout       int
}

var (
	RaumserverDebug = false
	MongoDbDrop     = true
	MongoDb         = "goweck"
	MongoUri        = "mongodb://192.168.2.1:27017"
	ListenPort      = 8080
	CheckInterval   = 5
	Database        *mgo.Database
	AlarmZoneId     = "uuid:C43C1A1D-AED1-472B-B0D0-210B7925000E"
	AlarmActive     = false
	NetClient       = &http.Client{Timeout: time.Second * 10}
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
		"id":    AlarmZoneId,
		"value": alarm.RadioChannel,
	}
	err := raumfeldAction(alarm.RaumserverUrl, "/controller/loadUri", params)
	return err
}

func adjustRaumfeldVolume(alarm *Alarm, volume int) error {
	params := map[string]string{
		"id":    AlarmZoneId,
		"scope": "zone",
		"value": fmt.Sprint(volume),
	}
	err := raumfeldAction(alarm.RaumserverUrl, "/controller/setVolume", params)
	return err
}

func stopRaumfeldStream(alarm *Alarm) error {
	params := map[string]string{
		"id": AlarmZoneId,
	}
	err := raumfeldAction(alarm.RaumserverUrl, "/controller/stop", params)
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
		t := time.Now()
		err := c.Find(bson.M{"hourminute": fmt.Sprintf("%02d%02d", t.Hour(), t.Minute())}).One(&result)
		if err != nil && strings.Contains(err.Error(), "not found") {
			fmt.Println("no alarm found")
		} else {
			if AlarmActive == false {
				fmt.Println("executing alarm")
				AlarmActive = true
				go executeAlarm(&result)
			} else {
				fmt.Println("alarm is already active")
			}
		}
	}
}

func main() {
	// initialize database
	session, err := mgo.Dial(GetEnvVar("MONGODB_URI", MongoUri))
	defer session.Close()
	Fatal(err)

	session.SetMode(mgo.Monotonic, true)

	if MongoDbDrop {
		err = session.DB(MongoDb).DropDatabase()
		Log(err)
	}
	Database = session.DB(MongoDb)

	ensureIndices()

	saveAlarm("0600", true, false, "BRF")
	saveAlarm("2239", true, false, "http://mp3channels.webradio.rockantenne.de/alternative")

	// execute or change alarms periodically
	go pollAlarm()

	// http endpoint for dealing with alarms
	mux := bone.New()
	mux.Get("/alarm", http.HandlerFunc(AlarmGetHandler))
	mux.Post("/alarm", http.HandlerFunc(AlarmPostHandler))
	mux.Handle("/", http.HandlerFunc(DocumentationHandler))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", ListenPort), mux))
}

//func ensureIndices(db *mgo.Database) {
func ensureIndices() {
	c := Database.C("alarm").With(Database.Session.Copy())

	index := mgo.Index{
		Key:        []string{"hourminute"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err := c.EnsureIndex(index)
	Fatal(err)
}

func saveAlarm(hourMinute string, weekDays bool, weekEnds bool, radioChannel string) {
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.Insert(&Alarm{
		HourMinute:    hourMinute,
		CreatedAt:     time.Now(),
		WeekDays:      weekDays,
		WeekEnds:      weekEnds,
		RaumserverUrl: "http://qnap:3535/raumserver",
		RadioChannel:  radioChannel,
		AlarmZone:     "all",
		VolumeStart:   20,
		VolumeEnd:     40,
		VolumeIncInt:  10,
		Timeout:       3600,
	})
	Fatal(err)
}

func AlarmGetHandler(w http.ResponseWriter, r *http.Request) {
	var results []Alarm
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.Find(bson.M{}).All(&results)
	Fatal(err)
	for _, alarm := range results {
		e, err := json.Marshal(alarm)
		Log(err)
		fmt.Fprintf(w, "%q", string(e))
	}
}

func AlarmPostHandler(w http.ResponseWriter, r *http.Request) {
}

func DocumentationHandler(w http.ResponseWriter, r *http.Request) {
}

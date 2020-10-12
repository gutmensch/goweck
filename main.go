package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	"github.com/gregdel/pushover"
	"github.com/gutmensch/goweck/app"
	"github.com/gutmensch/goweck/raumserver"
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
	ZoneUUID      string        `bson:"zoneUuid,omitempty"      json:"zoneUuid,omitempty"`
	ZoneName      string        `bson:"zoneName,omitempty"      json:"zoneName,omitempty"`
	StreamName    string        `bson:"streamName,omitempty"    json:"streamName,omitempty"`
	VolumeStart   int           `bson:"volumeStart,omitempty"   json:"volumeStart,omitempty"`
	VolumeEnd     int           `bson:"volumeEnd,omitempty"     json:"volumeEnd,omitempty"`
	VolumeIncStep int           `bson:"volumeIncStep,omitempty" json:"volumeIncStep,omitempty"`
	VolumeIncInt  int           `bson:"volumeIncInt,omitempty"  json:"volumeIncInt,omitempty"`
	Timeout       int           `bson:"timeout,omitempty"       json:"timeout,omitempty"`
}

var (
	Debug, _             = strconv.ParseBool(app.GetEnvVar("DEBUG", "false"))
	MongoDbDrop, _       = strconv.ParseBool(app.GetEnvVar("MONGODB_DROP", "false"))
	MongoDb              = app.GetEnvVar("MONGODB_DATABASE", "goweck")
	MongoURI             = app.GetEnvVar("MONGODB_URI", "mongodb://127.0.0.1:27017")
	Listen               = app.GetEnvVar("LISTEN", ":8080")
	CheckInterval        = 5
	Database             *mgo.Database
	enableDeepStandby, _ = strconv.ParseBool(app.GetEnvVar("DEEP_STANDBY", "false"))
	TimeZone             = app.GetEnvVar("TZ", "UTC")
	PushOverUser         = app.GetEnvVar("PUSHOVER_USER_TOKEN", "undefined")
	PushOverApp          = app.GetEnvVar("PUSHOVER_APP_TOKEN", "undefined")
	AlarmActive          = false
	DefaultRadioStreams  = []stream{
		stream{
			Name: "Rock Antenne",
			URL:  "http://mp3channels.webradio.rockantenne.de/alternative",
		},
		stream{
			Name: "Radio 21 Rock",
			URL:  "http://stream.radio21.de/radio21.mp3",
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
		stream{
			Name: "Oldies but Goldies",
			URL:  "http://mp3channels.webradio.antenne.de/oldies-but-goldies",
		},
		stream{
			Name: "MPD",
			URL:  "http://qnaps:8800",
		},
	}
)

func getStreamURL(name string) string {
	var result stream
	c := Database.C("stream").With(Database.Session.Copy())
	err := c.Find(bson.M{"name": name}).One(&result)
	if err != nil {
		return "not found"
	} else {
		return result.URL
	}
}

func sendFallbackMessage(msg string) error {
	app := pushover.New(PushOverApp)
	recipient := pushover.NewRecipient(PushOverUser)
	message := pushover.NewMessageWithTitle(msg, "GoWeck")
	response, err := app.SendMessage(message, recipient)
	if Debug {
		fmt.Println(response)
	}
	return err
}

func increaseVolume(alarm *alarm) {
	steps := alarm.VolumeEnd - alarm.VolumeStart
	for i := 1; i <= steps; i++ {
		if !AlarmActive {
			return
		}
		if Debug {
			fmt.Println("adjusting volume to", alarm.VolumeStart+i)
		}
		err := raumserver.AdjustRaumfeldVolume(alarm.ZoneUUID, alarm.VolumeStart+i)
		app.Log(err)
		time.Sleep(time.Duration(alarm.VolumeIncInt) * time.Second)
	}
}

func fallbackToPushover(alarm *alarm) {
	errCount := 1
	errCountEscalate := 1
	for {
		if !AlarmActive {
			return
		}
		running, _ := raumserver.CheckTransportState(alarm.ZoneUUID)
		if !running {
			errCount++
		}
		if (errCount % 10) == 0 {
			errCountEscalate++
			err := raumserver.PlayRaumfeldStream(alarm.ZoneUUID, getStreamURL(alarm.StreamName))
			app.Log(err)
		}
		if (errCountEscalate % 5) == 0 {
			err := sendFallbackMessage("Wake up, please!")
			app.Log(err)
			errCount = 1
			errCountEscalate = 1
		}
		time.Sleep(1 * time.Second)
	}
}

func startAlarm(alarm *alarm) {
	// play stream for wake up
	//if enableDeepStandby {
	err := raumserver.LeaveStandby(alarm.ZoneUUID)
	//}
	go increaseVolume(alarm)
	err = raumserver.PlayRaumfeldStream(alarm.ZoneUUID, getStreamURL(alarm.StreamName))
	app.Log(err)

	// send pushover notification as fallback
	go fallbackToPushover(alarm)
}

func stopAlarm(alarm *alarm) {
	err := raumserver.StopRaumfeldStream(alarm.ZoneUUID)
	app.Log(err)
	err = raumserver.AdjustRaumfeldVolume(alarm.ZoneUUID, alarm.VolumeStart)
	app.Log(err)
	if enableDeepStandby {
		err = raumserver.EnterStandby(alarm.ZoneUUID)
	}
}

func executeAlarm(alarm *alarm) {
	startAlarm(alarm)
	for i := alarm.Timeout; i > 0; i-- {
		if !AlarmActive {
			break
		}
		time.Sleep(1 * time.Second)
	}
	AlarmActive = false
	stopAlarm(alarm)
}

func pollAlarm() {
	var result alarm
	c := Database.C("alarm").With(Database.Session.Copy())

	for {
		<-time.After(time.Duration(CheckInterval) * time.Second)

		if AlarmActive {
			fmt.Println("[pollAlarm] alarm is currently active")
			continue
		}

		loc, err := time.LoadLocation(TimeZone)
		if err != nil {
			fmt.Printf("Error loading time zone %s with error %s\n", TimeZone, err.Error())
			continue
		}
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

		err = c.Find(search).One(&result)
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
	session, err := mgo.DialWithTimeout(MongoURI, 5*time.Second)
	app.Fatal(err)
	defer session.Close()

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

	// pushover test when starting
	go sendFallbackMessage("GoWeck starting up...")

	// http endpoint for dealing with alarms
	router := httpRouter()
	fmt.Printf("Listening on %s.\n", Listen)
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
	for _, stream := range DefaultRadioStreams {
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
	r.HandleFunc("/alarm/running", http.HandlerFunc(alarmRunningHandler)).Methods("GET")
	r.HandleFunc("/alarm/update/{id:[0-9a-f]{24}}", http.HandlerFunc(alarmUpdateHandler)).Methods("POST")
	r.HandleFunc("/alarm/delete/{id:[0-9a-f]{24}}", http.HandlerFunc(alarmDeleteHandler)).Methods("DELETE")
	r.HandleFunc("/", http.HandlerFunc(indexHandler)).Methods("GET")
	r.HandleFunc("/index.html", http.HandlerFunc(indexHandler)).Methods("GET")
	return r
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	data, _ := Asset("asset/index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func alarmListHandler(w http.ResponseWriter, r *http.Request) {
	var results []alarm
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.Find(bson.M{}).Sort("status", "hourMinute").All(&results)
	app.Fatal(err)
	e, err := json.Marshal(results)
	app.Log(err)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%q\n", string(e))
}

func alarmRunningHandler(w http.ResponseWriter, r *http.Request) {
	if AlarmActive {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func zoneListHandler(w http.ResponseWriter, r *http.Request) {
	zones, err := raumserver.GetZoneListJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%q\n", zones)
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
	AlarmActive = false
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
		StreamName:    "NOT_SET",
		ZoneUUID:      "",
		ZoneName:      raumserver.GetZoneName(new.ZoneUUID),
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

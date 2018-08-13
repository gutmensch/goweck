package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	//"os"
	"time"
	"strings"
	//"strconv"

        . "github.com/gutmensch/goweck/appbase"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/go-zoo/bone"
)

type Alarm struct {
	ID            bson.ObjectId `bson:"_id,omitempty"`
	FourDigitTime string
	CreatedAt     time.Time
        WeekDays      bool
        WeekEnds      bool
        RaumserverUrl string
        AlarmZone     string
        RadioChannel  string
        VolumeStart   int
        VolumeEnd     int
        TimeoutSec    int
}

var (
	MongoDbDrop = true
	MongoDb  = "goweck"
	MongoUri = "mongodb://192.168.2.1:27017"
	ListenPort = 8080
	CheckInterval = 2
        Database *mgo.Database
)

func executeAlarm(alarm *Alarm) {
	fmt.Println(alarm.RadioChannel)
}

func pollAlarm() {
	var result Alarm
	c := Database.C("alarm").With( Database.Session.Copy() )

	for {
		<-time.After(time.Duration(CheckInterval) * time.Second)
                t := time.Now()
		err := c.Find(bson.M{"fourdigittime": fmt.Sprintf("%02d%02d", t.Hour(), t.Minute())}).One(&result)
                if err != nil && strings.Contains(err.Error(), "not found") {
			fmt.Println("no alarm found")
                } else {
			go executeAlarm(&result)
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
	saveAlarm("0700", true, false, "RockAntenne")
	saveAlarm("1639", true, false, "RockAntenne")

	var result Alarm
	c := Database.C("alarm").With( Database.Session.Copy() )
	err = c.Find(bson.M{"radiochannel": "BRF"}).One(&result)
        Log(err)

	fmt.Println("Result: ", result)

	// execute or change alarms periodically
	go pollAlarm()

	// http endpoint for dealing with alarms
	mux := bone.New()

	mux.Get("/release/:id", http.HandlerFunc(ReleaseGetHandler))
	mux.Post("/release", http.HandlerFunc(ReleasePostHandler))

	mux.Handle("/", http.HandlerFunc(DocumentationHandler))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", ListenPort), mux))
}

//func ensureIndices(db *mgo.Database) {
func ensureIndices() {
	c := Database.C("alarm").With( Database.Session.Copy() )

	index := mgo.Index{
		Key:        []string{"fourdigittime"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err := c.EnsureIndex(index)
        Fatal(err)
}

func saveAlarm(fourDigitTime string, weekDays bool, weekEnds bool, radioChannel string) {
    c := Database.C("alarm").With( Database.Session.Copy() )
    err := c.Insert(&Alarm{
	FourDigitTime: fourDigitTime,
	CreatedAt: time.Now(),
	WeekDays: weekDays,
	WeekEnds: weekEnds,
	RaumserverUrl: "http://qnap:3535/raumserver",
	RadioChannel: radioChannel,
	AlarmZone: "all",
	VolumeStart: 30,
	VolumeEnd: 50,
	TimeoutSec: 3600,
    })
    Fatal(err)
}

func checkActiveAlarm() {
}

func ReleaseGetHandler(w http.ResponseWriter, r *http.Request) {
}

func ReleasePostHandler(w http.ResponseWriter, r *http.Request) {
}

func DocumentationHandler(w http.ResponseWriter, r *http.Request) {
}

package datastore

import (
	"errors"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gutmensch/goweck/internal/common"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

var alarmCollectionName = strings.ToLower(fmt.Sprintf("%T", Alarm{}))
var streamCollectionName = strings.ToLower(fmt.Sprintf("%T", Stream{}))

type MongoDatastore struct {
	URI      string
	Database *mgo.Database
	Session  *mgo.Session
	Timeout  time.Duration
	Debug    bool
}

func (m MongoDatastore) ListAlarms() []Alarm {
	return nil
}
func (m MongoDatastore) ListZones() []interface{} {
	return nil
}
func (m MongoDatastore) ListStreams() []Stream {
	return nil
}
func (m MongoDatastore) CreateAlarm() error {
	return nil
}
func (m MongoDatastore) UpdateAlarm() error {
	return nil
}
func (m MongoDatastore) DeleteAlarm() error {
	return nil
}
func (m MongoDatastore) StopAlarm() error {
	return nil
}
func (m MongoDatastore) GetCurrentAlarm(t time.Time) *Alarm {
	var result Alarm
	var err error

	c := m.Database.C(alarmCollectionName).With(m.Database.Session.Copy())
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
	debug := common.GetEnvVar("DEBUG")
	if err == mgo.ErrNotFound || err != nil {
		if Debug {
			fmt.Println(err.Error())
		}
		continue
	}
	return nil
}

func (m MongoDatastore) Init(dropOnInit bool) error {
	var err error

	m.Session, err = mgo.DialWithTimeout(m.URI, m.Timeout)
	m.Session.SetMode(mgo.Monotonic, true)

	if dropOnInit {
		err = m.Session.DB("").DropDatabase()
		if err != nil {
			common.Log.Info("error dropping database", zap.Error(err))
		}
	}
	m.Database = m.Session.DB("")

	m.ensureIndices()

	return nil
}

func (m MongoDatastore) Shutdown() error {
	m.Session.Close()
	return nil
}

func (m MongoDatastore) ensureIndices() {
	var err error
	var db = m.Database

	alarmCollection := db.C(alarmCollectionName).With(db.Session.Copy())
	index := mgo.Index{
		Key:        []string{"status", "hourMinute", "weekDays", "weekEnds"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	if err = alarmCollection.EnsureIndex(index); err != nil {
		common.Log.Info("error creating index", zap.Error(err))
	}

	streamCollection := db.C(streamCollectionName).With(db.Session.Copy())
	index = mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	if err = streamCollection.EnsureIndex(index); err != nil {
		common.Log.Info("error creating index", zap.Error(err))
	}
}

func (m MongoDatastore) SaveDefaultStreams(streams []Stream) error {
	var result Stream
	for _, stream := range streams {
		c := m.Database.C(streamCollectionName).With(m.Database.Session.Copy())
		dupStream := c.Find(bson.M{
			"name": stream.Name,
		}).One(&result)
		if !errors.Is(dupStream, mgo.ErrNotFound) {
			continue
		} else {
			err := c.Insert(stream)
			if err != nil {
				common.Log.Info("error saving stream into collection", zap.Error(err))
			}
		}
	}
}

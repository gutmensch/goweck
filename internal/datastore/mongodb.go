package datastore

import (
	"github.com/globalsign/mgo"
	"github.com/gutmensch/goweck/internal/common"
	"time"
)

type MongoDatastore struct {
	URI      string
	Database *mgo.Database
	Timeout time.Duration
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
func (m MongoDatastore) GetCurrentAlarm() *Alarm {
	return nil
}

func (m MongoDatastore) Init() {
	session, err := mgo.DialWithTimeout(m.URI, m.Timeout)
	common.Fatal(err)
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	if MongoDbDrop {
		err = session.DB(MongoDb).DropDatabase()
		common.Log(err)
	}
	Database = session.DB(MongoDb)

	ensureIndices()

	populateStreams()
}

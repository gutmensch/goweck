package datastore

import (
	"time"
)

type Datastore interface {
	ListAlarms() []Alarm
	ListZones() []interface{}
	ListStreams() []Stream
	CreateAlarm() error
	UpdateAlarm() error
	DeleteAlarm() error
	StopAlarm() error
	GetCurrentAlarm() *Alarm
}

type Config struct {
	DropOnInit bool
	Type       string
	URI        string
}

func New(cfg Config) Datastore {
	switch cfg.Type {
	// only MongoDB supported atm
	default:
		datastore := &MongoDatastore{
			URI: cfg.URI,
			Timeout: 5 * time.Second,
		}
		datastore.Init()
		return datastore
	}
}

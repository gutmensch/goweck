package datastore

import (
	"github.com/gutmensch/goweck/internal/common"
	"strconv"
	"time"
)

type Datastore interface {
	Init(bool) error
	Shutdown() error
	ListAlarms() []Alarm
	ListZones() []interface{}
	ListStreams() []Stream
	CreateAlarm() error
	UpdateAlarm() error
	DeleteAlarm() error
	StopAlarm() error
	GetCurrentAlarm(time.Time) *Alarm
	SaveDefaultStreams([]Stream) error
}

type Config struct {
	DropOnInit bool
	Type       string
	URI        string
}

func New(cfg Config) Datastore {

	debug, _ := strconv.ParseBool(common.GetEnvVar("DEBUG", "false"))

	switch cfg.Type {
	// only MongoDB supported atm
	default:
		datastore := &MongoDatastore{
			URI:     cfg.URI,
			Timeout: 5 * time.Second,
			Debug:   debug,
		}
		_ = datastore.Init(cfg.DropOnInit)
		return datastore
	}
}

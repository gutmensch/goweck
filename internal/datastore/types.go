package datastore

import "github.com/globalsign/mgo/bson"

type Stream struct {
	ID   bson.ObjectId `bson:"_id,omitempty"    json:"_id,omitempty"`
	Name string        `bson:"name,omitempty"   json:"name,omitempty"`
	URL  string        `bson:"url,omitempty"    json:"url,omitempty"`
}

type Alarm struct {
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

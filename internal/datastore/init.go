package datastore

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gutmensch/goweck/internal/common"
)

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
	common.Fatal(err)

	c = Database.C("stream").With(Database.Session.Copy())
	index = mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(index)
	common.Fatal(err)
}

func populateStreams() {
	var result dto.Stream
	for _, stream := range dto.DefaultRadioStreams {
		c := Database.C("stream").With(Database.Session.Copy())
		dupStream := c.Find(bson.M{
			"name": stream.Name,
		}).One(&result)
		if dupStream != mgo.ErrNotFound {
			continue
		} else {
			err := c.Insert(stream)
			if err != nil {
				common.Log(err)
			}
		}
	}
}

package api

import (
	"dario.cat/mergo"
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
	"github.com/gutmensch/goweck/internal/common"
	"github.com/gutmensch/goweck/internal/raumserver"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func AlarmListHandler(w http.ResponseWriter, r *http.Request) {
	var results []alarm
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.Find(bson.M{}).Sort("status", "hourMinute").All(&results)
	common.Fatal(err)
	e, err := json.Marshal(results)
	common.Log(err)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%q\n", string(e))
}

func AlarmRunningHandler(w http.ResponseWriter, r *http.Request) {
	if AlarmActive {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func ZoneListHandler(w http.ResponseWriter, r *http.Request) {
	zones, err := raumserver.GetZoneListJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%q\n", zones)
}

func StreamListHandler(w http.ResponseWriter, r *http.Request) {
	var results []stream
	c := Database.C("stream").With(Database.Session.Copy())
	err := c.Find(bson.M{}).All(&results)
	common.Fatal(err)
	e, err := json.Marshal(results)
	common.Log(err)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%q\n", string(e))
}

func AlarmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	common.Fatal(err)
	var new, old alarm
	err = json.Unmarshal(body, &new)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	common.Fatal(err)
	vars := mux.Vars(r)
	c := Database.C("alarm").With(Database.Session.Copy())
	err = c.FindId(bson.ObjectIdHex(vars["id"])).One(&old)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = mergo.Merge(&new, old)
	common.Log(err)
	err = c.UpdateId(bson.ObjectIdHex(vars["id"]), &new)
	if err != nil {
		common.Log(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func AlarmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	c := Database.C("alarm").With(Database.Session.Copy())
	err := c.RemoveId(bson.ObjectIdHex(vars["id"]))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func AlarmStopHandler(w http.ResponseWriter, r *http.Request) {
	AlarmActive = false
	w.WriteHeader(http.StatusOK)
}

func AlarmCreateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	common.Fatal(err)
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
		common.Log(err)
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

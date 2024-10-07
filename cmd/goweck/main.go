package main

import (
	"context"
	"fmt"
	"github.com/gutmensch/goweck/internal/common"
	"github.com/gutmensch/goweck/internal/datastore"
	"github.com/gutmensch/goweck/internal/http"
	"github.com/gutmensch/goweck/internal/job"
	"github.com/gutmensch/goweck/internal/raumserver"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gregdel/pushover"
)

var (
	Debug, _ = strconv.ParseBool(common.GetEnvVar("DEBUG", "false"))

	Database             *mgo.Database
	enableDeepStandby, _ = strconv.ParseBool(common.GetEnvVar("DEEP_STANDBY", "false"))
	TimeZone             = common.GetEnvVar("TZ", "UTC")
	PushOverUser         = common.GetEnvVar("PUSHOVER_USER_TOKEN", "undefined")
	PushOverApp          = common.GetEnvVar("PUSHOVER_APP_TOKEN", "undefined")
	AlarmActive          = false
)

func getStreamURL(name string) string {
	var result datastore.Stream
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

func increaseVolume(alarm *datastore.Alarm) {
	steps := alarm.VolumeEnd - alarm.VolumeStart
	for i := 1; i <= steps; i++ {
		if !AlarmActive {
			return
		}
		if Debug {
			fmt.Println("adjusting volume to", alarm.VolumeStart+i)
		}
		err := raumserver.AdjustRaumfeldVolume(alarm.ZoneUUID, alarm.VolumeStart+i)
		common.Log(err)
		time.Sleep(time.Duration(alarm.VolumeIncInt) * time.Second)
	}
}

func fallbackToPushover(alarm *datastore.Alarm) {
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
			common.Log(err)
		}
		if (errCountEscalate % 5) == 0 {
			err := sendFallbackMessage("Wake up, please!")
			common.Log(err)
			errCount = 1
			errCountEscalate = 1
		}
		time.Sleep(1 * time.Second)
	}
}

func startAlarm(alarm *datastore.Alarm) {
	// play stream for wake up
	//if enableDeepStandby {
	err := raumserver.LeaveStandby(alarm.ZoneUUID)
	//}
	go increaseVolume(alarm)
	err = raumserver.PlayRaumfeldStream(alarm.ZoneUUID, getStreamURL(alarm.StreamName))
	common.Log(err)

	// send pushover notification as fallback
	go fallbackToPushover(alarm)
}

func stopAlarm(alarm *datastore.Alarm) {
	err := raumserver.StopRaumfeldStream(alarm.ZoneUUID)
	common.Log(err)
	err = raumserver.AdjustRaumfeldVolume(alarm.ZoneUUID, alarm.VolumeStart)
	common.Log(err)
	if enableDeepStandby {
		err = raumserver.EnterStandby(alarm.ZoneUUID)
	}
}

func main() {
	var checkInterval time.Duration
	var dropOnInit bool
	var listen, timeZone string
	var err error

	// context and shutdown handling on signals
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		cancel()
	}()

	if dropOnInit, err = strconv.ParseBool(common.GetEnvVar("MONGODB_DROP", "false")); err != nil {
		common.Log.Error("error parsing drop on init", zap.Error(err))
		os.Exit(1)
	}
	ds := datastore.New(datastore.Config{
		Type:       "mongodb",
		URI:        common.GetEnvVar("MONGODB_URI", "mongodb://127.0.0.1:27017/goweck"),
		DropOnInit: dropOnInit,
	})

	// concurrent processes
	group, _ := errgroup.WithContext(ctx)

	listen = common.GetEnvVar("LISTEN", ":8080")
	server := http.NewServer(listen)
	group.Go(func() error {
		common.Log.Info("starting http server")
		return server.Run(ctx)
	})

	timeZone = common.GetEnvVar("TZ", "UTC")
	if checkInterval, err = time.ParseDuration(common.GetEnvVar("CHECK_INTERVAL", "10s")); err != nil {
		common.Log.Error("error parsing check interval", zap.Error(err))
		os.Exit(1)
	}
	alarmWatcher := &job.AlarmWatcher{
		Datastore:     ds,
		TimeZone:      timeZone,
		CheckInterval: checkInterval,
	}
	group.Go(func() error {
		common.Log.Info("starting alarmwatcher", zap.Duration("interval_s", checkInterval))
		return alarmWatcher.Run(ctx)
	})

	// pushover test when starting
	go sendFallbackMessage("GoWeck starting up...")

	// wait for processes
	if err = group.Wait(); err != nil {
		common.Log.Error("error waiting for child processes", zap.Error(err))
	}
}

package job

import (
	"context"
	"fmt"
	"github.com/gutmensch/goweck/internal/common"
	"github.com/gutmensch/goweck/internal/datastore"
	"go.uber.org/zap"
	"time"
)

type AlarmWatcher struct {
	Datastore     datastore.Datastore
	TimeZone      string
	CheckInterval time.Duration
	AlarmActive   bool
}

func (a AlarmWatcher) Run(ctx context.Context) error {
	// trigger shutdown when context done
	go func() {
		<-ctx.Done()
		common.Log.Info("terminating mongodb session")
		a.Datastore.Shutdown()
	}()
	return nil
}

func executeAlarm(alarm *datastore.Alarm) {
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

func (a AlarmWatcher) pollAlarm() {

	for {
		<-time.After(a.CheckInterval)

		if a.AlarmActive {
			common.Log.Info("alarm already active, skipping iteration")
			continue
		}

		loc, err := time.LoadLocation(a.TimeZone)
		if err != nil {
			common.Log.Info("error loading time zone", zap.Error(err))
			continue
		}
		t := time.Now().In(loc)

		query := a.Datastore.GetCurrentAlarm(t)
		if query == nil {
			continue
		}

		fmt.Println("[pollAlarm] alarm for current time and day found, executing.")
		AlarmActive = true
		go executeAlarm(&result)
	}
}

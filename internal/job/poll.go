package job

import (
	"context"
	"github.com/gutmensch/goweck/internal/datastore"
	"time"
)

type AlarmWatcher struct {
	Datastore     datastore.Datastore
	TimeZone      string
	CheckInterval time.Duration
}

func (a AlarmWatcher) Run(ctx context.Context) error {
	return nil
}

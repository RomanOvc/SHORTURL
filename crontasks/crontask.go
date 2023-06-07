package crontasks

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-co-op/gocron"
)

func RunCronJob(db *sql.DB) {
	s := gocron.NewScheduler(time.UTC)
	ctx := context.TODO()
	// s.Cron("0 1 * * *").Do(func() {
	// 	AddCountVisitOnIURLPerDay(db, ctx)
	// })

	// for test
	s.Cron("* * * * *").Do(func() {
		AddCountVisitOnIURLPerDay(db, ctx)
	})
	s.StartBlocking()

}

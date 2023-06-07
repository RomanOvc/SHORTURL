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
<<<<<<< HEAD
	// s.Cron("0 1 * * *").Do(func() {
	// 	AddCountVisitOnIURLPerDay(db, ctx)
	// })

	// for test
	s.Cron("* * * * *").Do(func() {
		AddCountVisitOnIURLPerDay(db, ctx)
=======
	s.Cron("0 1 * * *").Do(func() {
		addCountVisitOnIURLPerDay(db, ctx)
>>>>>>> c6beaf47ed1cd0b1c97992c7177a18a291ce2299
	})

	s.StartBlocking()
}

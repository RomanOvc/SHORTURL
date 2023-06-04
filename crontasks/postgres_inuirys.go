package crontasks

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"
)

type visitStatistic struct {
	Shorturl          string
	CountUniqueVisits string
	CountAllVisits    string
}

func addCountVisitOnIURLPerDay(db *sql.DB, ctx context.Context) {
	// FIXME передать контекст и транзаксия
	dateYesterday := time.Now().Add(-24 * time.Hour)
	dateStrat := fmt.Sprint(dateYesterday.Format("2006-01-02"))

	dateStop := fmt.Sprint(time.Now().Format("2006-01-02"))

	log.Println(dateStrat)
	log.Println(dateStop)
	rows, err := db.QueryContext(ctx, `
	SELECT 
		count(distinct useragent) as count_unique_visits, 
		count(useragent) as count_vists, 
		shorturl 
	FROM activity 
	WHERE click_time BETWEEN $1 AND $2 
	GROUP BY shorturl  
	ORDER BY shorturl ASC;
	`, dateStrat, dateStop)
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()

	var visitOnUrl []visitStatistic
	for rows.Next() {
		var v visitStatistic
		err := rows.Scan(&v.CountUniqueVisits, &v.CountAllVisits, &v.Shorturl)
		if err != nil {
			log.Panic(err)
		}

		visitOnUrl = append(visitOnUrl, v)
	}
	if err = rows.Err(); err != nil {
		log.Panic(err)
	}

	for _, g := range visitOnUrl {
		log.Println(g)
		db.QueryRowContext(ctx, "INSERT INTO visit_statistic(count_unique_visit, count_all_visit , index_url , date) values ($1,$2,$3,$4)", g.CountUniqueVisits, g.CountAllVisits, g.Shorturl, time.Now().Format("2006-01-02 15:04:05"))
		log.Println("данные добавлены")
	}

}

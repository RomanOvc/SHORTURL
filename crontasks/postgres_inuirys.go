package crontasks

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"
)

type visitStatistic struct {
	Shorturl          string `json:"shorturl"`
	CountUniqueVisits int    `json:"count_unique_visits"`
	CountAllVisits    int    `json:"count_all_visits"`
	ShorturlId        int    `json:"shorturl_id"`
}

func AddCountVisitOnURLPerDay(db *sql.DB, ctx context.Context) {
	dateYesterday := time.Now().Add(-24 * time.Hour)
	dateStrat := fmt.Sprint(dateYesterday.Format("2006-01-02"))
	dateStop := fmt.Sprint(time.Now().Format("2006-01-02"))

	// добавить порсент или долю  уникальынх визитов сколько% визитов было уникальных от общего числа посещения
	rows, err := db.QueryContext(ctx, `SELECT count(distinct useragent) as count_unique_visits, count(useragent) as count_vists,  shorturl, shorturl_id FROM activity WHERE click_time BETWEEN $1 AND $2 GROUP BY shorturl,shorturl_id ORDER BY shorturl ASC;`, dateStop, dateStrat)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var visitOnUrl []visitStatistic

	for rows.Next() {
		var v visitStatistic

		err := rows.Scan(&v.CountUniqueVisits, &v.CountAllVisits, &v.Shorturl, &v.ShorturlId)
		if err != nil {
			log.Fatal(err)
		}

		visitOnUrl = append(visitOnUrl, v)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	for _, g := range visitOnUrl {
		res, err := db.ExecContext(ctx, "INSERT INTO visit_statistic(count_unique_visit, count_all_visit , index_url ,shorturl_id, date) values ($1,$2,$3,$4,$5)", g.CountUniqueVisits, g.CountAllVisits, g.Shorturl, g.ShorturlId, time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			log.Fatal(err)
		}
		res.RowsAffected()
	}
}

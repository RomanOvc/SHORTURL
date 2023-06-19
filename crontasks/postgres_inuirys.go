package crontasks

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	pq "github.com/lib/pq"
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
	rows, err := db.QueryContext(ctx, `SELECT  count(distinct useragent) as count_unique_visits, count(useragent) as count_vists,  shorturl, shorturl_id FROM activity WHERE click_time BETWEEN $1 AND $2 GROUP BY shorturl,shorturl_id ORDER BY shorturl ASC;`, dateStrat, dateStop)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var count_unique_visit []int
	var count_all_visit []int
	var index_url []string
	var shorturl_id []int

	for rows.Next() {
		var v visitStatistic

		err := rows.Scan(&v.CountUniqueVisits, &v.CountAllVisits, &v.Shorturl, &v.ShorturlId)
		if err != nil {
			log.Fatal(err)
		}

		count_unique_visit = append(count_unique_visit, v.CountUniqueVisits)
		count_all_visit = append(count_all_visit, v.CountAllVisits)
		index_url = append(index_url, v.Shorturl)
		shorturl_id = append(shorturl_id, v.ShorturlId)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	res, err := db.ExecContext(ctx, `INSERT INTO  visit_statistic (count_unique_visit, count_all_visit , index_url ,shorturl_id) 
	SELECT unnest($1::int[]), unnest($2::int[]), unnest($3::text[]) ,unnest($4::int[])`, pq.Array(count_unique_visit), pq.Array(count_all_visit), pq.Array(index_url), pq.Array(shorturl_id))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)

}

func DeleteExpiredUid(db *sql.DB, ctx context.Context) {
	err := db.QueryRowContext(ctx, `delete from emailactivate WHERE active_until < $1`, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Fatal(err)
	}
}

// func Lol(db *sql.DB, ctx context.Context) {
// 	dateYesterday := time.Now().Add(24 * time.Hour)
// 	dateStrat := fmt.Sprint(dateYesterday.Format("2006-01-02"))
// 	dateStop := fmt.Sprint(time.Now().Format("2006-01-02"))

// 	_, err := db.ExecContext(ctx, `INSERT INTO visit_statistic(count_unique_visit, count_all_visit , index_url ,shorturl_id ) SELECT  count(distinct useragent) as count_unique_visits, count(useragent) as count_vists,  shorturl, shorturl_id FROM activity WHERE click_time BETWEEN $1 AND $2 GROUP BY shorturl,shorturl_id ORDER BY shorturl ASC;`, dateStop, dateStrat)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

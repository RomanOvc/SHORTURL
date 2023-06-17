package crontasks

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"
)

// FIXME почему экспортируемое?
type VisitStatistic struct {
	Shorturl          string `json:"shorturl"`
	CountUniqueVisits int    `json:"count_unique_visits"`
	CountAllVisits    int    `json:"count_all_visits"`
	ShorturlId        int    `json:"shorturl_id"`
}

// FIXME drop "I"
func AddCountVisitOnIURLPerDay(db *sql.DB, ctx context.Context) {
	dateYesterday := time.Now().Add(-24 * time.Hour)
	dateStrat := fmt.Sprint(dateYesterday.Format("2006-01-02"))

	dateStop := fmt.Sprint(time.Now().Format("2006-01-02"))

	log.Println(dateStrat)
	log.Println(dateStop)
	// добавить порсент или долю  уникальынх визитов сколько% визитов было уникальных от общего числа посещения
	rows, err := db.QueryContext(ctx, `SELECT count(distinct useragent) as count_unique_visits, count(useragent) as count_vists,  shorturl, shorturl_id FROM activity WHERE click_time BETWEEN $1 AND $2 GROUP BY shorturl,shorturl_id ORDER BY shorturl ASC;`, dateStop, dateStrat)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var visitOnUrl []VisitStatistic

	for rows.Next() {
		var v VisitStatistic

		err := rows.Scan(&v.CountUniqueVisits, &v.CountAllVisits, &v.Shorturl, &v.ShorturlId)
		if err != nil {
			log.Fatal(err)
		}

		visitOnUrl = append(visitOnUrl, v)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	// FIXME это уже продуктовая версия дропни ненужные для сопровождения логи
	log.Println(visitOnUrl)
	// unnest (вставка несколикх объектов)
	// подход к select изменить
	for _, g := range visitOnUrl {
		res, err := db.ExecContext(ctx, "INSERT INTO visit_statistic(count_unique_visit, count_all_visit , index_url ,shorturl_id, date) values ($1,$2,$3,$4,$5)", g.CountUniqueVisits, g.CountAllVisits, g.Shorturl, g.ShorturlId, time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			log.Println(g)
			log.Fatal(err)
		}
		//  сколько получилось с res сравнивать с количестов, которое я добавил. Если неравенство, записать в логах ошибку
		res.RowsAffected()
	}
}

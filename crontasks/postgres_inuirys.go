package crontasks

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"
)

<<<<<<< HEAD
type VisitStatistic struct {
	Shorturl          string `json:"shorturl"`
	CountUniqueVisits int    `json:"count_unique_visits"`
	CountAllVisits    int    `json:"count_all_visits"`
=======
type visitStatistic struct {
	Shorturl          string
	CountUniqueVisits string
	CountAllVisits    string
>>>>>>> c6beaf47ed1cd0b1c97992c7177a18a291ce2299
}

func addCountVisitOnIURLPerDay(db *sql.DB, ctx context.Context) {
	// FIXME передать контекст и транзаксия
	dateYesterday := time.Now().Add(24 * time.Hour)
	dateStrat := fmt.Sprint(dateYesterday.Format("2006-01-02"))

	dateStop := fmt.Sprint(time.Now().Format("2006-01-02"))

	log.Println(dateStrat)
	log.Println(dateStop)
<<<<<<< HEAD
	// добавить порсент или долю  уникальынх визитов сколько% визитов было уникальных от общего числа посещения
	rows, err := db.QueryContext(ctx, `SELECT count(distinct useragent) as count_unique_visits, count(useragent) as count_vists, shorturl FROM activity WHERE click_time BETWEEN $1 AND $2 GROUP BY shorturl ORDER BY shorturl ASC;`, dateStop, dateStrat)
=======
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
>>>>>>> c6beaf47ed1cd0b1c97992c7177a18a291ce2299
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var visitOnUrl []visitStatistic
	for rows.Next() {
		var v visitStatistic
		err := rows.Scan(&v.CountUniqueVisits, &v.CountAllVisits, &v.Shorturl)
		if err != nil {
			log.Fatal(err)
		}

		visitOnUrl = append(visitOnUrl, v)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	log.Println(visitOnUrl)
	// unnest (вставка несколикх объектов)
	// подход к select изменить
	for _, g := range visitOnUrl {
		res, err := db.ExecContext(ctx, "INSERT INTO visit_statistic(count_unique_visit, count_all_visit , index_url , date) values ($1,$2,$3,$4)", g.CountUniqueVisits, g.CountAllVisits, g.Shorturl, time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			log.Println(g)
			log.Fatal(err)
		}
		//  сколько получилось с res сравнивать с количестов, которое я добавил. Если неравенство, записать в логах ошибку
		res.RowsAffected()
	}

}

// сделать cascade tables

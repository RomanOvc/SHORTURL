package repository

import (
	"appurl/models"
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

type InquirysRepository struct {
	db *sql.DB
}

func NewInquirysRepository(db *sql.DB) *InquirysRepository {
	return &InquirysRepository{db: db}
}

func (r *InquirysRepository) AddGenerateUrl(ctx context.Context, shorturl, url, userEmail string) (string, error) {
	var (
		userId      int
		err         error
		tx          *sql.Tx
		shorturlRes string
	)
	tx, err = r.db.BeginTx(ctx, nil)

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// FIXME удлаить это нахуи
	err = tx.QueryRowContext(ctx, `SELECT user_id from users where useremail=$1`, userEmail).Scan(&userId)
	if err != nil {
		return "", errors.Wrap(err, "select err: user is empty")
	}

	err = tx.QueryRowContext(ctx, `INSERT INTO shortedurl (shorturl, originalurl,user_id) VALUES ($1, $2, $3) RETURNING shorturl ;`, shorturl, url, userId).Scan(&shorturlRes)
	if err != nil {
		return "", errors.Wrap(err, "select err: user is empty")
	}

	tx.Commit()

	return shorturlRes, nil
}

func (r *InquirysRepository) SelectShortUrlCount(shorturl string) (int, error) {
	var counter int

	err := r.db.QueryRow(`SELECT count(shorturl) FROM shortedurl where shorturl = $1`, shorturl).Scan(&counter)
	if err != nil {
		return 0, errors.Wrap(err, "SelectShortUrlCount()")
	}

	return counter, nil
}

func (r *InquirysRepository) SelectOriginalUrl(shorturl string) (*models.InfoUrl, error) {
	var infourl models.InfoUrl

	ro := r.db.QueryRow(`SELECT shorturl_id,originalurl FROM shortedurl where shorturl = $1`, shorturl)
	err := ro.Scan(&infourl.ShorturlId, &infourl.OriginalUrl)
	if err != nil {
		return nil, errors.Wrap(err, "SelectOriginalUrl()")
	}

	return &infourl, nil
}

func (r *InquirysRepository) SelectShortUrl(originalUrl string) (string, error) {
	var shorturl string

	err := r.db.QueryRow(`SELECT shorturl FROM shortedurl where originalurl = $1 `, originalUrl).Scan(&shorturl)
	switch err {
	case sql.ErrNoRows:
		return "", nil
	case nil:
		return shorturl, nil
	default:
		return "", err
	}
}

func (r *InquirysRepository) SelectUrlsByUser(ctx context.Context, useremail string) (*[]models.UrlsByUserStruct, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT 
		    s.shorturl, s.originalurl 
		FROM 
		    shortedurl AS s 
		JOIN users AS u 
		    ON u.user_id = s.userid 
		Where 
		    u.usermail=$1`,
		useremail)
	if err != nil {
		return nil, errors.Wrap(err, "no user_id")
	}

	defer rows.Close()

	var urlsByUserStruct []models.UrlsByUserStruct

	for rows.Next() {
		var uS models.UrlsByUserStruct

		err := rows.Scan(&uS.ShortUrl, &uS.OriginUrl)
		if err != nil {
			return nil, errors.Wrap(err, "no user_id")
		}

		urlsByUserStruct = append(urlsByUserStruct, uS)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "no rows")
	}

	return &urlsByUserStruct, nil

}

func (r *InquirysRepository) AddActivityInfo(ctx context.Context, shortUrl, userAgent, userPlatform string, shorturl_id int) (int, error) {
	var (
		activityId int
	)

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO activity (shorturl, useragent, platform, shorturl_id, click_time) 
		VALUES ($1, $2, $3, $4, $5) RETURNING activity_id;`,
		shortUrl, userAgent, userPlatform, shorturl_id, time.Now()).Scan(&activityId)
	if err != nil {
		return 0, errors.Wrap(err, "error add")
	}

	return activityId, errors.Wrap(err, "insert error")
}

func (r *InquirysRepository) VisitStatistic(ctx context.Context, shortUrl string) ([]models.VisitOnUrl, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT 
		    platform,count(platform) as count 
		FROM 
		    activity 
		where 
		    shorturl=$1 
		group by 
		    platform`,
		shortUrl)
	if err != nil {
		return nil, errors.Wrap(err, "error select")
	}

	defer rows.Close()

	var visitOnUrl []models.VisitOnUrl

	for rows.Next() {
		var v models.VisitOnUrl

		err := rows.Scan(&v.Platform, &v.Count)
		if err != nil {
			return nil, errors.Wrap(err, "error add into mass")
		}

		visitOnUrl = append(visitOnUrl, v)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error add")
	}

	return visitOnUrl, nil

}

func (r *InquirysRepository) CountVisitOnURL(ctx context.Context, url string) (int, error) {
	var countVisit int

	err := r.db.QueryRowContext(ctx, "select count(activity_id) from activity where shorturl=$1", url).Scan(&countVisit)
	if err != nil {
		return 0, errors.Wrap(err, "error add")
	}

	return countVisit, nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type InquirysInterface interface {
	AddGenerateUrl(shortUrl, url string) (string, error)
	SelectShortUrlCount(shortUrl string) (int, error)
	SelectOriginalUrl(shortUrl string) (string, error)
}

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

	err = tx.QueryRowContext(ctx, "SELECT user_id from users where usermail=$1", userEmail).Scan(&userId)
	if err != nil {
		return "", fmt.Errorf("select err: user is empty: %w", err)
	}

	err = tx.QueryRowContext(ctx, "INSERT INTO shortedurl (shorturl, originalurl,userid) VALUES ($1, $2, $3) RETURNING shorturl ;", shorturl, url, userId).Scan(&shorturlRes)
	if err != nil {
		return "", fmt.Errorf("insert error: %w", err)
	}
	tx.Commit()
	return shorturlRes, err
}

func (r *InquirysRepository) SelectShortUrlCount(shorturl string) (int, error) {
	var counter int
	err := r.db.QueryRow("SELECT count(shorturl) FROM shortedurl where shorturl = $1", shorturl).Scan(&counter)
	if err != nil {
		return 0, errors.Wrap(err, "repository/inquirys  SelectShortUrlCount() method error")
	}
	return counter, err
}

func (r *InquirysRepository) SelectOriginalUrl(shorturl string) (string, error) {
	var originalUrl string
	err := r.db.QueryRow("SELECT originalurl FROM shortedurl where shorturl = $1 ", shorturl).Scan(&originalUrl)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return originalUrl, err
}

func (r *InquirysRepository) SelectShortUrl(originalUrl string) (string, error) {
	var shorturl string
	err := r.db.QueryRow("SELECT shorturl FROM shortedurl where originalurl = $1 ", originalUrl).Scan(&shorturl)
	if err != nil {
		return "", errors.Wrap(err, "repository/inquirys  SelectOriginalUrl() method error")
	}
	return shorturl, err
}

// TODO
//

type UrlsByUserStruct struct {
	OriginUrl string `json:"origin_url"`
	ShortUrl  string `json:"short_url"`
}

func (r *InquirysRepository) SelectUrlsByUser(ctx context.Context, useremail string) (*[]UrlsByUserStruct, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT s.shorturl, s.originalurl FROM shortedurl AS s JOIN users AS u ON u.user_id = s.userid Where u.usermail=$1", useremail)
	if err != nil {
		return nil, errors.Wrap(err, "no user_id")
	}

	defer rows.Close()

	var urlsByUserStruct []UrlsByUserStruct
	for rows.Next() {
		var uS UrlsByUserStruct

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

// TODO добавление данных: юзер агент, коротка ссылка,
func (r *InquirysRepository) AddActivityInfo(ctx context.Context, shortUrl, userAgent, userPlatform string) (int, error) {
	var (
		activityId int
	)

	err := r.db.QueryRowContext(ctx, "INSERT INTO activity (shorturl, useragent, platform, click_time) VALUES ($1, $2, $3, $4) RETURNING activity_id;", shortUrl, userAgent, userPlatform, time.Now()).Scan(&activityId)
	if err != nil {
		return 0, errors.Wrap(err, "error add")
	}

	return activityId, errors.Wrap(err, "insert error")
}

type VisitOnUrl struct {
	Platform string `json:"platform"`
	Coint    int    `json:"count"`
}

func (r *InquirysRepository) VisitStatistic(ctx context.Context, shortUrl string) ([]VisitOnUrl, error) {
	rows, err := r.db.QueryContext(ctx, "select platform,count(platform)as count from activity where shorturl=$1 group by platform", shortUrl)
	if err != nil {
		return nil, errors.Wrap(err, "error select")
	}
	defer rows.Close()
	var visitOnUrl []VisitOnUrl
	for rows.Next() {
		var v VisitOnUrl
		err := rows.Scan(&v.Platform, &v.Coint)
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
	return countVisit, err
}

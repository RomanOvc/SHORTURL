package repository

import (
	"database/sql"

	"github.com/pkg/errors"
)

type InquirysInterface interface {
	AddGenerateUrl(shorturl, url string) (string, error)
	SelectShortUrlCount(shorturl string) (int, error)
	SelectOriginalUrl(shorturl string) (string, error)
}

type InquirysRepository struct {
	Db *sql.DB
}

func NewInquirysRepository(Db *sql.DB) *InquirysRepository {
	return &InquirysRepository{Db: Db}
}

func (r *InquirysRepository) AddGenerateUrl(shorturl, url string) (string, error) {
	_, err := r.Db.Exec("INSERT INTO shortedurl (shorturl, originalurl) VALUES ($1, $2);", shorturl, url)
	if err != nil {
		return "err", errors.Wrap(err, "repository/inquirys AddGenerateUrl() method error")
	}

	return "yes", err
}

func (r *InquirysRepository) SelectShortUrlCount(shorturl string) (int, error) {
	var counter int
	err := r.Db.QueryRow("SELECT count(shorturl) FROM shortedurl where shorturl = $1", shorturl).Scan(&counter)
	if err != nil {
		return 0, errors.Wrap(err, "repository/inquirys  SelectShortUrlCount() method error")
	}
	return counter, err
}

func (r *InquirysRepository) SelectOriginalUrl(shorturl string) (string, error) {
	var originalUrl string
	err := r.Db.QueryRow("SELECT originalurl FROM shortedurl where shorturl = $1", shorturl).Scan(&originalUrl)
	if err != nil {
		return "", errors.Wrap(err, "repository/inquirys  SelectOriginalUrl() method error")
	}
	return originalUrl, err
}

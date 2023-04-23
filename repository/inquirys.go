package repository

import (
	"database/sql"

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

func (r *InquirysRepository) AddGenerateUrl(shorturl, url string) (bool, error) {
	_, err := r.db.Exec("INSERT INTO shortedurl (shorturl, originalurl) VALUES ($1, $2);", shorturl, url)
	if err != nil {
		return false, errors.Wrap(err, "repository/inquirys AddGenerateUrl() method error")
	}

	return true, err
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
	err := r.db.QueryRow("SELECT originalurl FROM shortedurl where shorturl = $1", shorturl).Scan(&originalUrl)
	if err != nil {
		return "", errors.Wrap(err, "repository/inquirys  SelectOriginalUrl() method error")
	}
	return originalUrl, err
}

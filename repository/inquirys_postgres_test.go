package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSelectShortUrlCount(t *testing.T) {
	testcase := []struct {
		input  string
		mocka  func(sqlmock.Sqlmock)
		result int
	}{
		{
			input: "",
			mocka: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT count(shorturl) FROM shortedurl where shorturl = $1`)).
					WillReturnRows(sqlmock.NewRows([]string{"count(shorturl)"}).AddRow(0))
			},
			result: 0,
		},
		{
			input: "random string",
			mocka: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT count(shorturl) FROM shortedurl where shorturl = $1`)).
					WillReturnRows(sqlmock.NewRows([]string{"count(shorturl)"}).AddRow(1))
			},
			result: 1,
		},
	}

	for _, tc := range testcase {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("%d", err)
		}
		defer db.Close()

		tc.mocka(mock)

		res, _ := NewInquirysRepository(db).SelectShortUrlCount(tc.input)

		assert.Equal(t, tc.result, res, "%d != %d", tc.result, res)

	}

}

func TestSelectOriginalUrl(t *testing.T) {
	testcase := []struct {
		input  string
		mocka  func(sqlmock.Sqlmock)
		result string
	}{
		{
			input: "",
			mocka: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT originalurl FROM shortedurl where shorturl = $1`)).
					WillReturnRows(sqlmock.NewRows([]string{"originalurl"}).AddRow(""))
			},
			result: "",
		}, {
			input: "skldfjsdkfj",
			mocka: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT originalurl FROM shortedurl where shorturl = $1`)).
					WillReturnRows(sqlmock.NewRows([]string{"originalurl"}).AddRow("adasdasda"))
			},
			result: "adasdasda",
		},
	}

	for _, tc := range testcase {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("%d", err)
		}
		defer db.Close()

		tc.mocka(mock)

		res, _ := NewInquirysRepository(db).SelectOriginalUrl(tc.input)

		assert.Equal(t, tc.result, res, "%s != %s", tc.result, res)

	}

}

func TestAddGenerateUrl(t *testing.T) {
	testcase := []struct {
		shorturl    string
		originalurl string
		userId      int
		userEmail   string
		mocka       func(sqlmock.Sqlmock)
		result      string
	}{
		{
			shorturl:    "lLKHpIy",
			originalurl: "https://www.google.com/search?q=translate&oq=tran&aqs=chrome.1.69i57j35i39i650l2j0i512l2j69i65l3.6679j0j7&sourceid=chrome&ie=UTF-8",
			userId:      1,

			mocka: func(s sqlmock.Sqlmock) {

				s.ExpectBegin()

				s.ExpectQuery(regexp.QuoteMeta(`SELECT user_id from users where usermail=$1`)).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("1"))

				s.ExpectQuery(regexp.QuoteMeta(`INSERT INTO shortedurl (shorturl, originalurl,userid) VALUES ($1, $2, $3) RETURNING shorturl`)).
					WithArgs("lLKHpIy", "https://www.google.com/search?q=translate&oq=tran&aqs=chrome.1.69i57j35i39i650l2j0i512l2j69i65l3.6679j0j7&sourceid=chrome&ie=UTF-8", 1).
					WillReturnRows(sqlmock.NewRows([]string{"shorturl"}).AddRow("lLKHpIy"))

				s.ExpectCommit()
			},
			result: "lLKHpIy",
		},
		{
			shorturl:    "lLKHpIy",
			originalurl: "https://www.google.com/search?q=translate&oq=tran&aqs=chrome.1.69i57j35i39i650l2j0i512l2j69i65l3.6679j0j7&sourceid=chrome&ie=UTF-8",
			userId:      1,

			mocka: func(s sqlmock.Sqlmock) {

				s.ExpectBegin()

				s.ExpectQuery(regexp.QuoteMeta(`SELECT user_id from users where usermail=$1`)).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("1"))

				s.ExpectQuery(regexp.QuoteMeta(`INSERT INTO shortedurl (shorturl, originalurl,userid) VALUES ($1, $2, $3) RETURNING shorturl`)).
					WithArgs("asasdsda", "https://www.google.com/search?q=translate&oq=tran&aqs=chrome.1.69i57j35i39i650l2j0i512l2j69i65l3.6679j0j7&sourceid=chrome&ie=UTF-8", 1).
					WillReturnRows(sqlmock.NewRows([]string{"shorturl"}).AddRow(""))

				s.ExpectCommit()
			},
			result: "",
		},
	}

	for _, tc := range testcase {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("%d", err)
		}

		defer db.Close()

		tc.mocka(mock)

		res, _ := NewInquirysRepository(db).AddGenerateUrl(context.Background(), tc.shorturl, tc.originalurl, tc.userEmail)

		assert.Equal(t, tc.result, res, "%d != %d", tc.result, res)
	}
}

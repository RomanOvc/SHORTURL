package repository

import (
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
		mocka       func(sqlmock.Sqlmock)
		result      bool
	}{
		{
			shorturl:    "asd",
			originalurl: "asd",
			mocka: func(s sqlmock.Sqlmock) {
				s.ExpectExec(regexp.QuoteMeta(`INSERT INTO shortedurl (shorturl, originalurl) VALUES ($1, $2)`)).
					WithArgs("", "").WillReturnResult(sqlmock.NewResult(0, 0))
			},
			result: false,
		},
		{
			shorturl:    "asd",
			originalurl: "asd",
			mocka: func(s sqlmock.Sqlmock) {
				s.ExpectExec(regexp.QuoteMeta(`INSERT INTO shortedurl (shorturl, originalurl) VALUES ($1, $2)`)).
					WithArgs("asd", "asd").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			result: true,
		},
	}

	for _, tc := range testcase {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("%d", err)
		}
		defer db.Close()

		tc.mocka(mock)

		res, _ := NewInquirysRepository(db).AddGenerateUrl(tc.shorturl, tc.originalurl)

		assert.Equal(t, tc.result, res, "%d != %d", tc.result, res)
	}
}

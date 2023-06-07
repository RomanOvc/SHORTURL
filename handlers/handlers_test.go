package handlers

import (
	"appurl/repository"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

type testSet struct {
	token  string
	url    string
	mocka  func(sqlmock.Sqlmock)
	expect int
}

func TestCreateShortUrl(t *testing.T) {
	// os.Setenv("DOMAIN", "http://127.0.0.1:8000")

	tS := []testSet{
		{
			token: "",
			url:   "https://github.com/RomanOvc/SHORTURL/blob/unit_tests/handlers/handlers_test.go",
			mocka: func(s sqlmock.Sqlmock) {
				s.ExpectQuery(regexp.QuoteMeta(`SELECT count(shorturl) FROM shortedurl where shorturl = $1`)).WillReturnRows(sqlmock.NewRows([]string{"count(urlshort)"}).AddRow(0))
			},
			expect: http.StatusCreated,
		},
	}

	for _, testcase := range tS {

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("%d", err)

		}
		defer db.Close()

		testcase.mocka(mock)

		var buf bytes.Buffer

		if err := json.NewEncoder(&buf).Encode(UrlReqStruct{Url: testcase.url}); err != nil {
			t.Fatal("encode error")
		}

		req := httptest.NewRequest(http.MethodPost, "/take_larg_url", &buf,)
		w := httptest.NewRecorder()

		NewUseRepository(repository.NewInquirysRepository(db)).CreateShortUrl(w, req)

		assert.Equal(t, testcase.expect, w.Result().StatusCode, "%d != %d ", w.Result().StatusCode, testcase.expect)

	}
}

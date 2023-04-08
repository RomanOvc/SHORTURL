// bot будет получать запрос и возвращать ссылку
package handlers

import (
	"appurl/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

//custom error

func TestCreateShortUrl(t *testing.T) {
	os.Setenv("DOMAIN", "http://127.0.0.1:8000")

	for _, testcase := range []struct {
		input  string
		want   func(sqlmock.Sqlmock)
		expect int
	}{
		{
			input: "https://google.com",
			want: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT count").
					WillReturnRows(sqlmock.NewRows([]string{"count(shorturl)"}).AddRow(0))
			},
			expect: http.StatusCreated,
		},
		{
			input: "https://google.com",
			want: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT count").
					WillReturnRows(sqlmock.NewRows([]string{"count(shorturl)"}).AddRow(1))
			},
			expect: http.StatusOK,
		},
		{
			input:  "google.com",
			want:   func(m sqlmock.Sqlmock) {},
			expect: http.StatusBadRequest,
		},
		{
			input: "https://google.com",
			want: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT count").
					WillReturnError(fmt.Errorf("Пошел нахуй"))
			},
			expect: http.StatusBadRequest,
		},
	} {
		// 1 Mock db
		db, m, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		testcase.want(m)
		// 1

		// 2 Mock http
		var buf bytes.Buffer

		if err := json.NewEncoder(&buf).Encode(UrlReqStruct{Url: testcase.input}); err != nil {
			t.Fatal("encode error")
		}

		req := httptest.NewRequest(http.MethodPost, "/take_larg_url", &buf)
		w := httptest.NewRecorder()
		// 2

		NewUseRepository(repository.NewInquirysRepository(db)).CreateShortUrl(w, req)

		assert.Equal(t, testcase.expect, w.Result().StatusCode, "Test case ")
	}
}

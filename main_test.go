package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"database/sql/driver"

	"github.com/DATA-DOG/go-sqlmock"
)

type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestSample(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	{
		mock.ExpectExec(
			regexp.QuoteMeta(
				`INSERT INTO access_log (ip, access_ts, url_path) VALUES ($1, $2, $3)`,
			)).
			WithArgs(r.RemoteAddr, AnyTime{}, "/").
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	{
		columns := []string{"id", "ip", "access_ts", "url_path"}
		rows := sqlmock.NewRows(columns).
			AddRow("1", r.RemoteAddr, "/", time.Now())
		mock.ExpectQuery(`SELECT id, ip, url_path, access_ts FROM "access_log"`).
			WithArgs().
			WillReturnRows(rows)
	}

	rootHandlerGenerator(db)(w, r)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatal("unexpected status code")
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal("unexpected error")
	}
	const expected = `{"content_type":"","msg":"hello","a":"","b":""}` + "\n"
	if s := string(b); s != expected {
		t.Fatalf("unexpected response:\nexpected: %s\nactual: %s", s, expected)
	}
}

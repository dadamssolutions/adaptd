package adaptd

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var checkNumber int

func TestHTTPSRedirectHTTP(t *testing.T) {
	checkNumber = 0
	ts := httptest.NewServer(EnsureHTTPS(false)(http.HandlerFunc(handlerTester)))
	defer ts.Close()

	client := ts.Client()
	client.CheckRedirect = checkRedirect
	resp, err := client.Get(ts.URL)

	if err == nil || resp.StatusCode != http.StatusTemporaryRedirect || checkNumber != 0 {
		t.Error("HTTP request not redirected")
	}
}
func TestHTTPSRedirectHTTPS(t *testing.T) {
	checkNumber = 0
	ts := httptest.NewTLSServer(EnsureHTTPS(false)(http.HandlerFunc(handlerTester)))
	defer ts.Close()

	client := ts.Client()
	resp, err := client.Get(ts.URL)

	if err != nil || resp.StatusCode != http.StatusOK || checkNumber != 1 {
		log.Println(err)
		t.Error("HTTPS request unexpectedly redirected")
	}
}

func TestDisallowingLongerPathsBasic(t *testing.T) {
	checkNumber = 0
	server := httptest.NewServer(DisallowLongerPaths("/", http.HandlerFunc(http.NotFound))((http.HandlerFunc(handlerTester))))
	defer server.Close()
	client := server.Client()
	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK || checkNumber != 1 {
		t.Error("Standard get request should not return an error")
	}

	reqURL := server.URL + "/login"
	req, _ = http.NewRequest("GET", reqURL, nil)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusNotFound || checkNumber != 1 {
		t.Errorf("A request with URL %v to %v should produce an error\n", reqURL, server.URL)
	}
}

func TestDisallowingLongerPathsWithLongerURL(t *testing.T) {
	checkNumber = 0
	server := httptest.NewServer(DisallowLongerPaths("/login", http.HandlerFunc(http.NotFound))(http.HandlerFunc(handlerTester)))
	defer server.Close()
	server.URL += "/login"
	client := server.Client()
	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK || checkNumber != 1 {
		t.Error("Standard get request should not return an error")
	}

	reqURL := server.URL + "/user_id/12345"
	req, _ = http.NewRequest("GET", reqURL, nil)
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusNotFound || checkNumber != 1 {
		t.Errorf("A request with URL %v to %v should produce an error\n", reqURL, server.URL)
	}
}

func TestTxContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	checkNumber = 0
	server := httptest.NewServer(PutTxOnContext(db)(http.HandlerFunc(handlerTester)))
	defer server.Close()
	server.URL += "/login/"

	client := server.Client()
	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != http.StatusOK || checkNumber != 1 || mock.ExpectationsWereMet() != nil {
		t.Errorf("A simple request should begin transaction, call handler, and commit transaction")
	}
}

func TestTxPanic(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	checkNumber = 0
	server := httptest.NewServer(PutTxOnContext(db)(http.HandlerFunc(handlerPanic)))
	defer server.Close()
	server.URL += "/login/"

	client := server.Client()
	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != http.StatusInternalServerError || checkNumber != 0 || mock.ExpectationsWereMet() != nil {
		t.Errorf("If the handler panics, we should get an internal server error")
	}
}

func handlerTester(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling test request")
	checkNumber++
}

func handlerPanic(w http.ResponseWriter, r *http.Request) {
	panic("Panic should rollback")
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	return fmt.Errorf("Redirected to %v", req.URL)
}

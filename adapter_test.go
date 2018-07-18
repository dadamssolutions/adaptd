package adaptd

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var checkNumber int

func TestDisallowingLongerPathsBasic(t *testing.T) {
	checkNumber = 0
	server := httptest.NewServer(DisallowLongerPaths("/")((http.HandlerFunc(handlerTester))))
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
	server := httptest.NewServer(DisallowLongerPaths("/login")(http.HandlerFunc(handlerTester)))
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

func handlerTester(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling test request")
	checkNumber++
}

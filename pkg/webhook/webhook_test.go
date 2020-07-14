package webhook

import (
	"defaultallowpe/pkg/config"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppNotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/foobar", nil)

	config := config.New()
	app := New(config)
	res, _ := app.Test(req)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestApiNotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/vN/foobar", nil)

	config := config.New()
	app := New(config)
	res, _ := app.Test(req)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, res.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err.Error())
	}
	var resBody map[string]interface{}
	err = json.Unmarshal(bodyBytes, &resBody)
	if err != nil {
		t.Fatal("failed to json decode res body")
	}
	expected := http.StatusText(http.StatusNotFound)
	if expected != resBody["status"] {
		t.Errorf("expected %s, got %s", expected, resBody["status"])
	}
}

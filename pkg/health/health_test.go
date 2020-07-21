package health

import (
	"defaultallowpe/pkg/config"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber"
)

func TestHealthApi(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)

	config, _ := config.New()
	app := fiber.New()
	Routes(app.Group(""), config)
	res, _ := app.Test(req)

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, res.StatusCode)
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
	if !resBody["ready"].(bool) {
		t.Error("expected ready true, got ready false")
	}
}

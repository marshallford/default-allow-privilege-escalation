package webhook

import (
	"bytes"
	"defaultallowpe/pkg/config"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	admissionReviewCreatePod = admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: metav1.GroupVersionKind{
				Kind: "Pod",
			},
			Operation: "CREATE",
		},
	}
)

func TestMutateGetMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/mutate", nil)

	config := config.New()
	app := New(config)
	res, _ := app.Test(req)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestMutateBadContentType(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/mutate", nil)
	config := config.New()
	app := New(config)
	res, _ := app.Test(req)
	if res.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("expected status code %d, got %d", http.StatusUnsupportedMediaType, res.StatusCode)
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
	expectedError := "invalid content-type, expected application/json"
	if resBody["error"] != expectedError {
		t.Errorf("expected error message %s, got %s", expectedError, resBody["error"])
	}
}

func TestMutateEmptyAdmissionRequest(t *testing.T) {
	admissionReview := admissionv1.AdmissionReview{}
	admissionReview.TypeMeta = admissionReviewCreatePod.TypeMeta

	arBytes, err := json.Marshal(admissionReview)
	if err != nil {
		t.Fatal("failed to json encode AdmissionReview")
	}
	req := httptest.NewRequest("POST", "/api/v1/mutate", bytes.NewReader(arBytes))
	req.Header.Set("Content-Type", "application/json")

	config := config.New()
	app := New(config)
	res, _ := app.Test(req)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, res.StatusCode)
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
	expectedError := "unexpected nil AdmissionRequest"
	if resBody["error"] != expectedError {
		t.Errorf("expected error message %s, got %s", expectedError, resBody["error"])
	}
}

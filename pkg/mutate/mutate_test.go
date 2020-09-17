package mutate

import (
	"bytes"
	"defaultallowpe/pkg/config"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
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
	secret = corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-secret",
		},
		Data: map[string][]byte{
			"foo": []byte("bar"),
		},
	}
	containerNoSecurityContext = corev1.Container{
		Name:  "foo",
		Image: "image:tag",
	}
	containerSecurityContextEmpty = corev1.Container{
		Name:            "foo",
		Image:           "image:tag",
		SecurityContext: &corev1.SecurityContext{},
	}
	containerSecurityContextWithOtherField = corev1.Container{
		Name:  "foo",
		Image: "image:tag",
		SecurityContext: &corev1.SecurityContext{
			Privileged: func() *bool {
				b := true
				return &b
			}(),
		},
	}
	containerSecurityContextWithField = corev1.Container{
		Name:  "foo",
		Image: "image:tag",
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: func() *bool {
				b := true
				return &b
			}(),
		},
	}
)

func pod(ns string, initContainers []corev1.Container, containers []corev1.Container) corev1.Pod {
	return corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-pod",
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			InitContainers: initContainers,
			Containers:     containers,
		},
	}
}

func TestMutateBadAdmissionRequestObject(t *testing.T) {
	secretBytes, err := json.Marshal(secret)

	if err != nil {
		t.Fatal("failed json encode Secret")
	}

	tt := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "gibberish",
			input:    []byte("foobar"),
			expected: `couldn't get version/kind; json parse error: json: cannot unmarshal string into Go value of type struct { APIVersion string "json:\"apiVersion,omitempty\""; Kind string "json:\"kind,omitempty\"" }`,
		},
		{
			name:     "secret",
			input:    secretBytes,
			expected: "unexpected type *v1.Secret",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			admissionReview := admissionv1.AdmissionReview{}
			admissionReview.TypeMeta = admissionReviewCreatePod.TypeMeta
			admissionReview.Request = admissionReviewCreatePod.Request
			admissionReview.Request.Kind = admissionReviewCreatePod.Request.Kind
			admissionReview.Request.Object.Raw = tc.input
			res := mutate(&admissionReview, false)
			if res.Result.Message != tc.expected {
				t.Errorf("expected message %s, got %s", tc.expected, res.Result.Message)
			}
			if res.Result.Status != metav1.StatusFailure {
				t.Errorf("expected status %s, got %s", metav1.StatusFailure, res.Result.Status)
			}
		})
	}
}

func TestMutateNoPatches(t *testing.T) {
	tt := []struct {
		name     string
		input    corev1.Pod
		expected []patch
	}{
		{
			name:  "namespace system",
			input: pod(metav1.NamespaceSystem, []corev1.Container{}, []corev1.Container{containerNoSecurityContext}),
		},
		{
			name:  "namespace system",
			input: pod(metav1.NamespacePublic, []corev1.Container{}, []corev1.Container{containerNoSecurityContext}),
		},
		{
			name:  "container security context with field",
			input: pod("default", []corev1.Container{}, []corev1.Container{containerSecurityContextWithField}),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			podBytes, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatal("failed to json encode Pod")
			}

			admissionReview := admissionv1.AdmissionReview{}
			admissionReview.TypeMeta = admissionReviewCreatePod.TypeMeta
			admissionReview.Request = admissionReviewCreatePod.Request
			admissionReview.Request.Kind = admissionReviewCreatePod.Request.Kind
			admissionReview.Request.Object.Raw = podBytes
			res := mutate(&admissionReview, false)

			if res.Patch != nil {
				t.Errorf("expected no patch, got %s", res.Patch)
			}

			if res.PatchType != nil {
				t.Errorf("expected no patch type, got %v", res.PatchType)
			}
		})
	}
}

func TestMutatePatches(t *testing.T) {
	tt := []struct {
		name     string
		input    corev1.Pod
		expected []patch
	}{
		{
			name:  "container no security context",
			input: pod("default", []corev1.Container{}, []corev1.Container{containerNoSecurityContext}),
			expected: []patch{
				{
					Op:    "add",
					Path:  "/spec/containers/0/securityContext",
					Value: struct{}{},
				},
				{
					Op:    "add",
					Path:  "/spec/containers/0/securityContext/allowPrivilegeEscalation",
					Value: false,
				},
			},
		},
		{
			name:  "container empty security context",
			input: pod("default", []corev1.Container{}, []corev1.Container{containerSecurityContextEmpty}),
			expected: []patch{
				{
					Op:    "add",
					Path:  "/spec/containers/0/securityContext/allowPrivilegeEscalation",
					Value: false,
				},
			},
		},
		{
			name:  "container security context with other field",
			input: pod("default", []corev1.Container{}, []corev1.Container{containerSecurityContextWithOtherField}),
			expected: []patch{
				{
					Op:    "add",
					Path:  "/spec/containers/0/securityContext/allowPrivilegeEscalation",
					Value: false,
				},
			},
		},
		{
			name:  "initcontainer no security context",
			input: pod("default", []corev1.Container{containerNoSecurityContext}, []corev1.Container{}),
			expected: []patch{
				{
					Op:    "add",
					Path:  "/spec/initContainers/0/securityContext",
					Value: struct{}{},
				},
				{
					Op:    "add",
					Path:  "/spec/initContainers/0/securityContext/allowPrivilegeEscalation",
					Value: false,
				},
			},
		},
		{
			name:  "initcontainer empty security context",
			input: pod("default", []corev1.Container{containerSecurityContextEmpty}, []corev1.Container{}),
			expected: []patch{
				{
					Op:    "add",
					Path:  "/spec/initContainers/0/securityContext/allowPrivilegeEscalation",
					Value: false,
				},
			},
		},
		{
			name:  "initcontainer security context with other field",
			input: pod("default", []corev1.Container{containerSecurityContextWithOtherField}, []corev1.Container{}),
			expected: []patch{
				{
					Op:    "add",
					Path:  "/spec/initContainers/0/securityContext/allowPrivilegeEscalation",
					Value: false,
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			podBytes, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatal("failed to json encode Pod")
			}

			admissionReview := admissionv1.AdmissionReview{}
			admissionReview.TypeMeta = admissionReviewCreatePod.TypeMeta
			admissionReview.Request = admissionReviewCreatePod.Request
			admissionReview.Request.Kind = admissionReviewCreatePod.Request.Kind
			admissionReview.Request.Object.Raw = podBytes
			res := mutate(&admissionReview, false)

			expectedBytes, err := json.Marshal(tc.expected)
			if err != nil {
				t.Fatal("failed to json encode patch")
			}
			if !bytes.Equal(expectedBytes, res.Patch) {
				t.Errorf("expected patch %s, got %s", expectedBytes, res.Patch)
			}
		})
	}
}

func TestMutateApiFailures(t *testing.T) {
	secretBytes, err := json.Marshal(secret)
	if err != nil {
		t.Fatal("failed json encode Secret")
	}
	admissionReview := admissionv1.AdmissionReview{}
	admissionReview.TypeMeta = admissionReviewCreatePod.TypeMeta
	arBytes, err := json.Marshal(admissionReview)
	if err != nil {
		t.Fatal("failed to json encode AdmissionReview")
	}

	tt := []struct {
		name               string
		input              io.Reader
		setContentType     bool
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "content type",
			input:              nil,
			setContentType:     false,
			expectedStatusCode: http.StatusUnsupportedMediaType,
			expectedError:      "invalid content-type, expected application/json",
		},
		{
			name:               "bad content",
			input:              strings.NewReader("foobar"),
			setContentType:     true,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "could not decode AdmissionReview",
		},
		{
			name:               "unexpected resource",
			input:              bytes.NewReader(secretBytes),
			setContentType:     true,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "unexpected GroupVersionKind: /v1, Kind=Secret",
		},
		{
			name:               "empty request",
			input:              bytes.NewReader(arBytes),
			setContentType:     true,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "unexpected nil AdmissionRequest",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/mutate", tc.input)
			if tc.setContentType {
				req.Header.Set("Content-Type", "application/json")
			}

			config, _ := config.New()
			app := fiber.New()
			Routes(app.Group(""), config)
			res, _ := app.Test(req)

			if res.StatusCode != tc.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tc.expectedStatusCode, res.StatusCode)
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
			resError := resBody["error"]
			if resError != tc.expectedError {
				t.Errorf("expected error message %s, got %s", tc.expectedError, resError)
			}
		})
	}
}

func TestMutateApiSuccess(t *testing.T) {
	pod := pod("default", []corev1.Container{}, []corev1.Container{containerNoSecurityContext})
	podBytes, err := json.Marshal(pod)
	if err != nil {
		t.Fatal("failed to json encode Pod")
	}

	admissionReview := admissionv1.AdmissionReview{}
	admissionReview.TypeMeta = admissionReviewCreatePod.TypeMeta
	admissionReview.Request = admissionReviewCreatePod.Request
	admissionReview.Request.Kind = admissionReviewCreatePod.Request.Kind
	admissionReview.Request.Object.Raw = podBytes
	arBytes, err := json.Marshal(admissionReview)
	if err != nil {
		t.Fatal("failed to json encode AdmissionReview")
	}

	req := httptest.NewRequest("POST", "/mutate", bytes.NewReader(arBytes))
	req.Header.Set("Content-Type", "application/json")

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
	if !resBody["response"].(map[string]interface{})["allowed"].(bool) {
		t.Error("expected allowed true, got allowed false")
	}
	expected := "JSONPatch"
	patchType := resBody["response"].(map[string]interface{})["patchType"].(string)
	if patchType != expected {
		t.Errorf("expected patchType %s, got patchType %s", expected, patchType)
	}
}

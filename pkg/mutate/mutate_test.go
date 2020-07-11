package mutate

import (
	"bytes"
	"encoding/json"
	"testing"

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

func TestBadAdmissionRequestObject(t *testing.T) {
	secret, err := json.Marshal(corev1.Secret{
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
	})

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
			input:    secret,
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

func TestApiMutateNoPatches(t *testing.T) {
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

func TestApiMutatePatches(t *testing.T) {
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

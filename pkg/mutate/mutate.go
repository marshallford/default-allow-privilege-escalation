package mutate

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/spf13/viper"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	scheme       = runtime.NewScheme()
	codecs       = serializer.NewCodecFactory(scheme)
	deserializer = codecs.UniversalDeserializer()
)

type patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type appError struct {
	Error string `json:"error"`
}

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(admissionv1.AddToScheme(scheme))
}

// Handler is the HTTP handler for mutate requests
func Handler(config *viper.Viper, c *fiber.Ctx) {
	// validate Content-Type
	if !c.Is("json") {
		err := c.Status(fiber.StatusUnsupportedMediaType).JSON(&appError{
			Error: "invalid content-type, expected application/json",
		})
		if err != nil {
			c.Next(err)
		}
		return
	}

	// get AdmissionReview
	reviewGVK := admissionv1.SchemeGroupVersion.WithKind("AdmissionReview")
	obj, gvk, err := deserializer.Decode([]byte(c.Body()), &reviewGVK, &admissionv1.AdmissionReview{})
	if err != nil {
		err := c.Status(fiber.StatusBadRequest).JSON(&appError{
			Error: "could not decode AdmissionReview",
		})
		if err != nil {
			c.Next(err)
		}
		return
	}

	review, ok := obj.(*admissionv1.AdmissionReview)
	if !ok {
		err := c.Status(fiber.StatusBadRequest).JSON(&appError{
			Error: fmt.Sprintf("unexpected GroupVersionKind: %s", gvk),
		})
		if err != nil {
			c.Next(err)
		}
		return
	}

	// check if request is empty
	if review.Request == nil {
		err := c.Status(fiber.StatusBadRequest).JSON(&appError{
			Error: "unexpected nil AdmissionRequest",
		})
		if err != nil {
			c.Next(err)
		}
		return
	}

	// mutate
	admissionResponse := mutate(review, config.GetBool("app.default"))

	// return new AdmissionReview
	review.Response = admissionResponse
	review.Response.UID = review.Request.UID
	err = c.Status(fiber.StatusOK).JSON(review)
	if err != nil {
		c.Next(err)
	}
}

func mutationRequired(metadata *metav1.ObjectMeta) bool {
	ignoredNamespaces := []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
	for _, namespace := range ignoredNamespaces {
		if metadata.Namespace == namespace {
			return false
		}
	}
	return true
}

func patchContainer(basepath string, sc *corev1.SecurityContext, defaultAllowPrivilegeEscalation bool) []patch {
	var patches []patch
	if sc == nil {
		patches = append(patches, patch{
			Op:    "add",
			Path:  basepath,
			Value: corev1.SecurityContext{},
		})
	}

	if sc == nil || sc.AllowPrivilegeEscalation == nil {
		patches = append(patches, patch{
			Op:    "add",
			Path:  fmt.Sprintf("%v/allowPrivilegeEscalation", basepath),
			Value: defaultAllowPrivilegeEscalation,
		})
	}
	return patches
}

func mutate(ar *admissionv1.AdmissionReview, defaultAllowPrivilegeEscalation bool) *admissionv1.AdmissionResponse {
	obj, _, err := deserializer.Decode(ar.Request.Object.Raw, nil, nil)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
				Status:  metav1.StatusFailure,
			},
		}
	}

	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: fmt.Sprintf("unexpected type %T", obj),
				Status:  metav1.StatusFailure,
			},
		}
	}

	// check if mutation is required
	if !mutationRequired(&pod.ObjectMeta) {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// look for containers in pod to patch
	var patches []patch
	for i, c := range pod.Spec.InitContainers {
		path := fmt.Sprintf("/spec/initContainers/%v/securityContext", i)
		patches = append(patches, patchContainer(path, c.SecurityContext, defaultAllowPrivilegeEscalation)...)
	}

	for i, c := range pod.Spec.Containers {
		path := fmt.Sprintf("/spec/containers/%v/securityContext", i)
		patches = append(patches, patchContainer(path, c.SecurityContext, defaultAllowPrivilegeEscalation)...)
	}

	// allow request if there aren't any patches
	if len(patches) == 0 {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// encodes patches as json
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
				Status:  metav1.StatusFailure,
			},
		}
	}

	// respond with patches
	return &admissionv1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

package mutate

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
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

// Routes manages Fiber routes for mutate pkg
func Routes(r fiber.Router, config *viper.Viper) {
	r.Post("/mutate", HandlerFunc(config))
}

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(admissionv1.AddToScheme(scheme))
}

// HandlerFunc returns a func that is a HTTP handler for mutate requests
func HandlerFunc(config *viper.Viper) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// validate Content-Type
		if !c.Is("json") {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(&appError{
				Error: "invalid content-type, expected application/json",
			})
		}

		// get AdmissionReview
		reviewGVK := admissionv1.SchemeGroupVersion.WithKind("AdmissionReview")
		obj, gvk, err := deserializer.Decode(c.Body(), &reviewGVK, &admissionv1.AdmissionReview{})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&appError{
				Error: "could not decode AdmissionReview",
			})
		}

		review, ok := obj.(*admissionv1.AdmissionReview)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(&appError{
				Error: fmt.Sprintf("unexpected GroupVersionKind: %s", gvk),
			})
		}

		// check if request is empty
		if review.Request == nil {
			return c.Status(fiber.StatusBadRequest).JSON(&appError{
				Error: "unexpected nil AdmissionRequest",
			})
		}

		// mutate
		admissionResponse := mutate(review, config.GetBool("app.default"))

		// return new AdmissionReview
		review.Response = admissionResponse
		review.Response.UID = review.Request.UID
		return c.Status(fiber.StatusOK).JSON(review)
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

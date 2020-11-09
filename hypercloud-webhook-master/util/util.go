package util

import (
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//Jsonpatch를 담을 수 있는 구조체
type PatchOps struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// Jsonpatch를 하나 만들어서 slice에 추가하는 함수
func CreatePatch(po *[]PatchOps, o, p string, v interface{}) {
	*po = append(*po, PatchOps{
		Op:    o,
		Path:  p,
		Value: v,
	})
}

// Response.result.message에 err 메시지 넣고 반환
func ToAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

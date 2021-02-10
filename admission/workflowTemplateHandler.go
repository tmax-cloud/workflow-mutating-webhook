package admission

import (
	"fmt"

	"k8s.io/api/admission/v1beta1"
)

func WorkflowTemplateSACheck(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	fmt.Println("TODO for WorkflowTemplate")
	return nil
}

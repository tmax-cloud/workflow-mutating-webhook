package admission

import (
	"encoding/json"
	"fmt"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/klog"
)

func WorkflowSACheck(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}

	fmt.Println("check for enter metadata handler")

	ms := wfv1.Workflow{}
	ds := "default-editor"

	if err := json.Unmarshal(ar.Request.Object.Raw, &ms); err != nil {
		return ToAdmissionResponse(err) //msg: error
	}

	//수정사항을 담을 구조체 slice
	var patch []patchOps

	if len(ms.Spec.ServiceAccountName) == 0 && len(ms.Spec.WorkflowTemplateRef.Name) == 0 {
		createPatch(&patch, "add", "/spec/serviceAccountName", ds)
	}
	klog.Infof("check data for ms : %s", ms.Spec)

	// 구조체 slice에 저장된 patch를 []byte로 변경
	if patchData, err := json.Marshal(patch); err != nil {
		return ToAdmissionResponse(err) //msg: error
	} else {
		klog.Infof("JsonPatch=%s", string(patchData))
		reviewResponse.Patch = patchData
	}

	// v1beta1 pkg에 저장된 patchType (const string)을 Resp에 저장
	pt := v1beta1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	reviewResponse.Allowed = true

	return &reviewResponse
}

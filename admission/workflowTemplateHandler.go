package admission

import (
	"encoding/json"
	"fmt"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
)

func WorkflowTemplateSACheck(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}


	fmt.Println("check for enter metadata handler")
	
	ms := WorkflowTemplate{}
	
	ds := "default-editor"
	if err := json.Unmarshal(ar.Request.Object.Raw, &ms); err != nil {
		return ToAdmissionResponse(err) //msg: error
	}

	ws := &ms

	//수정사항을 담을 구조체 slice
	var patch []patchOps


	//fmt.Println("check data for ms : %s", ms)

	if len(ms.WorkflowSpec.ServiceAccountName) == 0 {
		createPatch(&patch, "add", "/spec/serviceAccountName", ds)
		if len(ms.WorkflowSpec.WorkflowTemplateRef.Name) == 0 {
			createPatch(&patch, "add", "/spec/serviceAccountName", ds)
		} else {
			if len(ws.Spec.ServiceAccountName) == 0{
				createPatch(&patch, "add", "/spec/serviceAccountName", ds)
			} 
			//templateRef를 타고 들어가서 확인, SA가 없으면
			//default-editor 추가
		}
	}
	fmt.Println("check data for ws : %s", ms.WorkflowSpec)

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

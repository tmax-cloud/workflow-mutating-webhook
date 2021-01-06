package admission

import (
	"encoding/json"
	//"time"
	"fmt"
	//jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	wfv1 "github.com/yuhanjung/argo/pkg/apis/workflow/v1alpha1"
	
)
type WorkflowTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`
	Spec	wfv1.WorkflowTemplateSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
}
func (wftmpl *WorkflowTemplate) GetTemplateByName(name string) *wfv1.Template {
	for _, t := range wftmpl.Spec.Templates {
		if t.Name == name {
			return &t
		}
	}
	return nil
}


// yaml을 담을 struct
type Meta struct {
	metav1.TypeMeta	`json:",inline"`                                                 // kind & apigroup
	metav1.ObjectMeta	`json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"` // metadata
	wfv1.WorkflowSpec	`json:"spec" protobuf:"bytes,2,opt,name=spec "`//spec
}

type WFMeta struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`
	wfv1.WorkflowTemplateSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
}

func AddResourceMeta(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}

	ms := Meta{}
	//ts := WFMeta{}
	//ws := &WorkflowTemplate{}
	
	ds := "default-editor"

	if err := json.Unmarshal(ar.Request.Object.Raw, &ms); err != nil {
		return ToAdmissionResponse(err) //msg: error
	}

	//수정사항을 담을 구조체 slice
	var patch []patchOps


	//fmt.Println("check data for ms : %s", ms)

	if len(ms.WorkflowSpec.ServiceAccountName) == 0 {
		createPatch(&patch, "add", "/spec/serviceAccountName", ds)
		if len(ms.WorkflowSpec.WorkflowTemplateRef.Name) == 0 {
			createPatch(&patch, "add", "/spec/serviceAccountName", ds)
		} else {
			ws.GetTemplateByName(ms.WorkflowSpec.WorkflowTemplateRef.Name)
			//if err := ws.GetTemplateByName(ms.WorkflowSpec.WorkflowTemplateRef.Name); err!= nil {}
			if len(ws.Spec.ServiceAccountName) == 0{
				createPatch(&patch, "add", "/spec/serviceAccountName", ds)
			} 
			//templateRef를 타고 들어가서 확인, SA가 없으면
			//default-editor 추가
		}
	}
	fmt.Println("check data for ws : %s", ms.WorkflowSpec)
	//fmt.Println("check data for ws : %s", ws)

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

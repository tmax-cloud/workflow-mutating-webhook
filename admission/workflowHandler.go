package admission

import (
	"encoding/json"
	"fmt"
	
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/klog"
	//"context"
	//sav1 "k8s.io/client-go/1.5/kubernetes/typed/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)



func WorkflowSACheck(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}

	fmt.Println("check for enter metadata handler")

	ms := wfv1.Workflow{}
	ds := "default-editor"

	if err := json.Unmarshal(ar.Request.Object.Raw, &ms); err != nil {
		return ToAdmissionResponse(err) //msg: error
	}

	/* serviceaccount client-go
	ctx := context.Background()
	sa := sav1.serviceAccounts{}
	//clicmd.GetDefaultServer()
	getopt := metav1.GetOptions{}
	sa.Get(ctx, "default-editor", getopt)*/



	//수정사항을 담을 구조체 slice
	var patch []patchOps

	a := 0
	if len(ms.Spec.ServiceAccountName) == 0 && ms.Spec.WorkflowTemplateRef == nil {
		klog.Infof("in if")
		if len(ms.Spec.Templates) == 0 { 
			klog.Infof("in if, len templates 0")
			createPatch(&patch, "add", "/spec/serviceAccountName", ds)
		} else {
			klog.Infof("in if, len templates != 0")
			for i := 0 ; i < len(ms.Spec.Templates); i++ {
				if len(ms.Spec.Templates[i].ServiceAccountName) == 0 {
					a = a+1
					//templates의 항목에 넣어주는 부분
					//templatestring := "/spec/templates[" + a + "]/serviceAccountName"
					//klog.Infof("check data for templatestring : %s", templatestring)
					//createPatch(&patch, "add", templatestring, ds)
				} 
			}
			if a > 0{
				createPatch(&patch, "add", "/spec/serviceAccountName", ds)
			}
		} 
	}
	//klog.Infof("check data for ms.Spec : %s", ms.Spec)
	
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

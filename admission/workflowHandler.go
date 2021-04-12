package admission

import (
	"context"
	"encoding/json"
	"fmt"

	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	//wfv1 "github.com/argoproj/argo-workflows/pkg/apis/workflow/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func WorkflowSACheck(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}

	fmt.Println("check for enter metadata handler")

	// serviceaccount client-go
	
	ds := "default-editor"
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil{
		panic(err.Error())
	}

	_, err = clientset.CoreV1().ServiceAccounts("argo").Get(context.TODO(), ds, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("default-editor isn't exist\n")

	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("default-editor is exist\n")
	}


	ms := wfv1.Workflow{}
    
	if err := json.Unmarshal(ar.Request.Object.Raw, &ms); err != nil {
		return ToAdmissionResponse(err) //msg: error
	}

	
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
					a  = a+1
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

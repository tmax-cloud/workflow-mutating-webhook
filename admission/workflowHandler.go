package admission

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	//"runtime/debug"
	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	//wfv1 "github.com/argoproj/argo-workflows/pkg/apis/workflow/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/klog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func WorkflowSACheck(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}

	fmt.Println("check for enter metadata handler")

	ms := wfv1.Workflow{}
    
	if err := json.Unmarshal(ar.Request.Object.Raw, &ms); err != nil {
		return ToAdmissionResponse(err) //msg: error
	}
	nsofworkflow := ms.ObjectMeta.Namespace

	klog.Infof("workflow created in namespace : %s", nsofworkflow)
		
	ds := "default-editor"
	config1, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	client1, err := kubernetes.NewForConfig(config1)
	if err != nil{
		panic(err.Error())
	}

	// argo namespace에서 default-editor 확인 후 없을 경우 생성
	_, err = client1.CoreV1().ServiceAccounts(nsofworkflow).Get(context.TODO(), ds, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("serviceaccont default-editor isn't exist\n")
		config2, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		client2, err := kubernetes.NewForConfig(config2)
		if err != nil {
			panic(err)
		}
		serviceAccountClient := client2.CoreV1().ServiceAccounts(nsofworkflow)

		serviceAccount := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
					Name: "default-editor",
			},	
		}
		result, err := serviceAccountClient.Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
		
		if err != nil {			
			panic(err)
		}
		klog.Infof("Created service account %v.", result.GetObjectMeta().GetName())

		config3, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		client3, err := kubernetes.NewForConfig(config3)
		if err != nil {
			panic(err)
		}
		RbacClient := client3.RbacV1().RoleBindings(nsofworkflow)

		_, err = RbacClient.Get(context.TODO(), ds, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			klog.Infof("rolebinding is not exist")

			saRole := &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
						Name: "default-editor",
						Namespace: nsofworkflow,
				},
				RoleRef: rbacv1.RoleRef{
					Kind: "ClusterRole",
					Name: "kubeflow-edit",
				},
				Subjects: []rbacv1.Subject {
					{
						Kind: "ServiceAccount",
						Name: "default-editor",
						Namespace: nsofworkflow,
					},
				},
			}
			rbac, err := RbacClient.Create(context.TODO(), saRole, metav1.CreateOptions{})
			if err != nil {			
				panic(err)
			}
			klog.Infof("role binding created name is : %v.", rbac.GetObjectMeta().GetName())

		}
		
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("default-editor is exist\n")
	}

	var patch []patchOps

	
	if len(ms.Spec.ServiceAccountName) == 0 && ms.Spec.WorkflowTemplateRef == nil {
		if len(ms.Spec.Templates) == 0 {
			createPatch(&patch, "add", "/spec/serviceAccountName", ds)
		} else {
			for i := 0 ; i < len(ms.Spec.Templates); i++ {
				if len(ms.Spec.Templates[i].ServiceAccountName) == 0 {
					//templates의 항목에 넣어주는 부분
					a := strconv.FormatInt(int64(i),10)
					templatestring := "/spec/templates/" + a +"/serviceAccountName"
					klog.Infof("check data for templatestring : %s", templatestring)
					createPatch(&patch, "add", templatestring, ds)
				}
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
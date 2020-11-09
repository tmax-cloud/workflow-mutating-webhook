package admission

import (
	"encoding/json"
	"time"

	util "hypercloud4-webhook/util"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// 사용자가 작성한 manifest를 담을 구조체 선언
type Meta struct {
	metav1.TypeMeta   `json:",inline"`                                                 // kind & apigroup
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"` // annotation
}

func AddResourceMeta(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	// Resp 선언
	reviewResponse := v1beta1.AdmissionResponse{}

	// Annotation에 넣을 created/updatedTime 구하기
	currentTime := time.Now()

	// Req에서 필요한 정보 파싱
	userName := ar.Request.UserInfo.Username
	operation := string(ar.Request.Operation)
	ms := Meta{}   // New object의 metadata를 저장하는 구조체 선언
	oms := Meta{}  // Old object의 metadata를 저장하는 구조체 선언
	diff := Meta{} // New와 Old의 차이를 저장하는 구조체 (meta를 변경하려고 하는지 확인하기 위해서)

	// New object를 ms 구조체에 저장
	if err := json.Unmarshal(ar.Request.Object.Raw, &ms); err != nil {
		return util.ToAdmissionResponse(err) //msg: error
	}
	// Old object가 존재하는지 확인 (Create action에서는 old object는 len == 0인 []byte)
	if len(ar.Request.OldObject.Raw) > 0 {
		// Old object가 존재하면 oms 구조체에 저장
		if err := json.Unmarshal(ar.Request.OldObject.Raw, &oms); err != nil {
			return util.ToAdmissionResponse(err) //msg: error
		}
		// New와 Old의 diff json을 계산
		if mergePatch, err := jsonpatch.CreateMergePatch(ar.Request.OldObject.Raw, ar.Request.Object.Raw); err != nil {
			return util.ToAdmissionResponse(err) //msg: error
		} else {
			// Diff json을 diff 구조체에 저장
			if err := json.Unmarshal(mergePatch, &diff); err != nil {
				return util.ToAdmissionResponse(err) //msg: error
			}
		}
	}

	// Meta를 직접 생성/수정하려는 경우 요청을 거절
	if denyReq(ms, diff, operation) {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: "Can not creat/update resource metadata.",
			},
		}
	}

	// Resp에 담을 jsonpatch 구조체의 slice 선언
	var patch []util.PatchOps

	// ms에 annotation 필드 존재 여부 확인
	if len(ms.Annotations) == 0 {
		// 없으면 annotation에 들어갈  key, value만들고
		am := map[string]interface{}{
			"creator":     userName,
			"createdTime": currentTime,
			"updater":     userName,
			"updatedTime": currentTime,
		}
		// 위의 key,value를 갖는 annotation 객체를 add하는 jsonpatch 생성
		util.CreatePatch(&patch, "add", "/metadata/annotations", am)
	} else {
		// annotation이 있으면, annotation 내부 key value만 생성
		// creator 없으면 생성
		if _, ok := ms.Annotations["creator"]; !ok {
			util.CreatePatch(&patch, "add", "/metadata/annotations/creator", userName)
		}
		// createdTime 없으면 생성
		if _, ok := ms.Annotations["createdTime"]; !ok {
			util.CreatePatch(&patch, "add", "/metadata/annotations/createdTime", currentTime)
		}
		// update는 무조건 생성
		util.CreatePatch(&patch, "add", "/metadata/annotations/updater", userName)
		util.CreatePatch(&patch, "add", "/metadata/annotations/updatedTime", currentTime)
	}

	// 구조체 slice에 저장된 patch를 []byte로 변경
	if patchData, err := json.Marshal(patch); err != nil {
		return util.ToAdmissionResponse(err) //msg: error
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

func denyReq(ms, diff Meta, op string) bool {
	// Create의 경우 annotation에 meta 있으면 mutating 거절 (거절 시 resource 생성 여부는 config에서 설정)
	if op == "CREATE" {
		if _, ok := ms.Annotations["creator"]; ok {
			return true
		} else if _, ok := ms.Annotations["createdTime"]; ok {
			return true
		} else if _, ok := ms.Annotations["updater"]; ok {
			return true
		} else if _, ok := ms.Annotations["updatedTime"]; ok {
			return true
		}
	}

	// Update의 경우 diff 구조체에 meta 있으면 deny
	if op == "UPDATE" {
		if _, ok := diff.Annotations["creator"]; ok {
			return true
		} else if _, ok := diff.Annotations["createdTime"]; ok {
			return true
		} else if _, ok := diff.Annotations["updater"]; ok {
			return true
		} else if _, ok := diff.Annotations["updatedTime"]; ok {
			return true
		}
	}

	return false
}

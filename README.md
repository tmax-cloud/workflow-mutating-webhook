# workflow-mutating-webhook
argo-workflow 리소스에 적절한 ServiceAccount를 맵핑하기 위한 웹훅
  - workflow / workflowtemplate 생성을 감지하다가, 생성요청이 오면 webhook하여 serviceaccount(sa) 항목이 있는지 확인
    - workflow의 경우, workflowtemplateRef 항목에 값이 들어있으면 통과
  - ref도 없고, sa도 없을 경우, workflow / workflowtemplate 생성 요청이 온 namespace의 sa 중 default-editor 가 있는지 확인
  - default-editor가 없으면 생성 및 rbac 적용
  - default-ediotr를 sa 항목에 적용
    - 이때 templates 의 각 항목중에 sa가 없는 항목이 있을 경우, 비어있는 해당 항목에만 sa 적용
  - workflow / workflowtemplate가 생성됨

# dependency
argo v2.12.11
cert-manager
issuer(cert-manager 동작을 위해)

# workflow-mutating-webhook
argo-workflow 리소스에 적절한 ServiceAccount를 맵핑하기 위한 웹훅

# build 방법
go build .

docker build -t image:tag . 

# certmanager 사용방법
certificate.yaml 파일을 apply 한다.
이때 namespace는 webhook server와 같이 하고,
사용할 secretName을 지정
issuerRef 필드의 경우, 현재 공용으로 사용중인 ck-selfsigned-clusterissuer 사용

mutatingwebhookconfiguration의 annotation에 다음과 같은 값을 넣어준다.
cert-manager.io/inject-ca-from: namespace/certificate name

webhook server deployment에 secret name 을 certificate에서 생성해준 secret으로 설정

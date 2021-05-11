# workflow-mutating-webhook
argo-workflow 리소스에 적절한 ServiceAccount를 맵핑하기 위한 웹훅입니다.

# build 방법
go build .

docker build -t image:tag . 

# https 인증서 mount 방법
$ openssl req -nodes -new -x509 -keyout ca.key -out ca.crt -subj "/CN=Admission Controller Webhook Demo CA"

$ openssl genrsa -out webhook-server-tls.key 2048

$ openssl req -new -key webhook-server-tls.key -subj "/CN=servicename.namespace.svc" \
    | openssl x509 -req -CA ca.crt -CAkey ca.key -CAcreateserial -out webhook-server-tls.crt

kubectl -n namespace create secret tls [mount 시킬 secret 이름] \
    --cert "webhook-server-tls.crt" \
    --key "webhook-server-tls.key"

kubectl apply -f wf-webhook-deployment.yaml -n namespace

$ export CA_PEM_BASE64="$(openssl base64 -A <"ca.crt")"

$ cat wf-webhook-configuration.yaml | sed "s/{{CA_PEM_BASE64}}/$CA_PEM_BASE64/g" | kubectl apply -n namespace -f -

# certmanager 사용방법
certificate.yaml 파일을 apply 한다.
이때 namespace는 webhook server와 같이 하고,
사용할 secretName을 지정
issuerRef 필드의 경우, 현재 공용으로 사용중인 ck-selfsigned-clusterissuer 사용

mutatingwebhookconfiguration의 annotation에 다음과 같은 값을 넣어준다.
cert-manager.io/inject-ca-from: namespace/certificate name

webhook server deployment에 secret name 을 certificate에서 생성해준 secret으로 설정

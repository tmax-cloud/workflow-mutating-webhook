# workflow-mutating-webhook
mutating webhook for workflow

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
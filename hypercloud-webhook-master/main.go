package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/klog"

	"encoding/json"

	"io/ioutil"

	"k8s.io/api/admission/v1beta1"

	admission "hypercloud4-webhook/admission"
	audit "hypercloud4-webhook/audit"
	util "hypercloud4-webhook/util"
)

type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	klog.Info(fmt.Sprintf("handling request: %s", body))

	requestedAdmissionReview := v1beta1.AdmissionReview{}
	responseAdmissionReview := v1beta1.AdmissionReview{}

	if err := json.Unmarshal(body, &requestedAdmissionReview); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = util.ToAdmissionResponse(err)
	} else {
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	respBytes, err := json.Marshal(responseAdmissionReview)

	klog.Infof("sending response: %s", respBytes)

	if err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = util.ToAdmissionResponse(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = util.ToAdmissionResponse(err)
	}
}

func serveMetadata(w http.ResponseWriter, r *http.Request) {
	serve(w, r, admission.AddResourceMeta)
}

func serveAudit(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		audit.GetAudit(w, r)
	case http.MethodPost:
		audit.AddAudit(w, r)
	case http.MethodPut:
	case http.MethodDelete:
	default:
		//error
	}
}

func serveAuditAuth(w http.ResponseWriter, r *http.Request) {
	audit.AddAuditAuth(w, r)
}

var (
	port     int
	certFile string
	keyFile  string
)

func main() {
	flag.IntVar(&port, "port", 8443, "hypercloud-webhook server port")
	flag.StringVar(&certFile, "certFile", "/run/secrets/tls/server.crt", "hypercloud-webhook server cert")
	flag.StringVar(&keyFile, "keyFile", "/run/secrets/tls/server.key", "x509 Private key file for TLS connection")
	flag.Parse()

	// crt와 key를 불러와서 변수에 저장
	keyPair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		klog.Errorf("Failed to load key pair: %s", err)
	}

	// URI에 맞는 handler 함수를 호출 (req multiplexer)
	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", serveMetadata)
	mux.HandleFunc("/audit", serveAudit)
	mux.HandleFunc("/audit/authentication", serveAuditAuth)

	// HTTPS 서버 설정
	whsvr := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),                              // 서버의 IP:PORT 지정
		Handler:   mux,                                                   // Req 처리할 handler 입력
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{keyPair}}, // 인증서 설정
	}

	klog.Info("Starting webhook server...")

	// HTTPS 서버 시작
	go func() {
		if err := whsvr.ListenAndServeTLS("", ""); err != nil { //HTTPS로 서버 시작
			klog.Errorf("Failed to listen and serve webhook server: %s", err)
		}
	}()

	go func() {
		for {
			if audit.Queue.Len() > 0 {
				items, _ := audit.Queue.Get(audit.Queue.Len())
				audit.InsertI(&items)
			}
			waitTime := time.NewTimer(time.Second * 10)
			<-waitTime.C
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	klog.Info("OS shutdown signal received...")
	whsvr.Shutdown(context.Background())
}

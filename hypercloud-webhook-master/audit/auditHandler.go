package audit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
)

func init() {
	Queue = queue.New(512)
}

var Queue *queue.Queue

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func AddAudit(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var eventList audit.EventList
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if err := json.Unmarshal(body, &eventList); err != nil {
		klog.Error(err)
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}
	insertE(&eventList.Items)
}

func GetAudit(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:tmax@tcp(mysql-service.hypercloud4-system.svc:3306)/audit?parseTime=true")
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	namespace := r.URL.Query().Get("namespace")
	resource := r.URL.Query().Get("resource")
	startTime := r.URL.Query().Get("startTime")
	endTime := r.URL.Query().Get("endTime")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")
	code := r.URL.Query().Get("code")
	sort := r.URL.Query()["sort"]

	var b strings.Builder
	b.WriteString("select * from audit.audit where 1=1 ")

	if startTime != "" && endTime != "" {
		b.WriteString("and stagetimestamp between '")
		b.WriteString(startTime)
		b.WriteString("' and '")
		b.WriteString(endTime)
		b.WriteString("' ")
	}

	if namespace != "" {
		b.WriteString("and namespace = '")
		b.WriteString(namespace)
		b.WriteString("' ")
	}

	if resource != "" {
		b.WriteString("and resource = '")
		b.WriteString(resource)
		b.WriteString("' ")
	}

	if code != "" {
		codeNum, _ := strconv.ParseInt(code, 10, 32)
		lowerBound := (codeNum / 100) * 100
		upperBound := lowerBound + 100
		b.WriteString("and code between '")
		b.WriteString(fmt.Sprintf("%v", lowerBound))
		b.WriteString("' and '")
		b.WriteString(fmt.Sprintf("%v '", upperBound))
	}

	if sort != nil && len(sort) > 0 {
		b.WriteString("order by ")
		for _, s := range sort {
			order := " asc, "
			if string(s[0]) == "-" {
				order = " desc, "
				s = s[1:]
			}
			b.WriteString(s)
			b.WriteString(order)
		}
		b.WriteString("stagetimestamp desc ")
	} else {
		b.WriteString("order by stagetimestamp desc ")
	}

	if limit != "" {
		b.WriteString("limit ")
		b.WriteString(limit)
		b.WriteString(" ")
	} else {
		b.WriteString("limit 100 ")
	}

	if offset != "" {
		b.WriteString("offset ")
		b.WriteString(offset)
		b.WriteString(" ")
	} else {
		b.WriteString("offset 0 ")
	}
	query := b.String()

	klog.Info("query: ", query)

	eventList := get(query)
	eventList.Kind = "EventList"
	eventList.APIVersion = "audit.k8s.io/v1"

	respBytes, err := json.Marshal(eventList)
	if err != nil {
		klog.Error(err)
	}
	klog.Infof("sending response: %s", respBytes)

	// json을 return
	if err != nil {
		klog.Error(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

func AddAuditAuth(w http.ResponseWriter, r *http.Request) {
	var body []byte
	var eventList audit.EventList
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if err := json.Unmarshal(body, &eventList); err != nil {
		klog.Error(err)
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	for _, item := range eventList.Items {
		item.AuditID = types.UID(uuid.New().String())
		item.ObjectRef = new(audit.ObjectReference)
		item.ObjectRef.APIGroup = AUTH_APIGROUP
		item.ObjectRef.APIVersion = AUTH_APIVERSION
		item.ObjectRef.Resource = AUTH_RESOURCE
		item.Stage = AUTH_STAGE
		item.StageTimestamp.Time = time.Now()

		if item.ResponseStatus.Code/100 == 2 {
			item.ResponseStatus.Status = "Success"
			item.ResponseStatus.Message = item.Verb + " sucess"
		} else {
			item.ResponseStatus.Status = "Failure"
			item.ResponseStatus.Message = item.Verb + " failed"
		}

		Queue.Put(item)
		if Queue.Len() >= 500 {
			items, _ := Queue.Get(Queue.Len())
			InsertI(&items)
		}

	}
}

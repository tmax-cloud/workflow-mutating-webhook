package audit

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
)

const (
	AUDIT_INSERT_QUERY = "insert into audit.audit values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	AUTH_APIGROUP      = "tmax.io"
	AUTH_APIVERSION    = "v1"
	AUTH_RESOURCE      = "users"
	AUTH_STAGE         = "ResponseComplete"
)

func InsertI(items *[]interface{}) {
	db, err := sql.Open("mysql", "root:tmax@tcp(mysql-service.hypercloud4-system.svc:3306)/audit?parseTime=true")
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	for _, item := range *items {
		event, _ := item.(audit.Event)
		_, err := db.Exec(AUDIT_INSERT_QUERY,
			event.AuditID,
			event.User.Username,
			event.UserAgent,
			event.ObjectRef.Namespace,
			event.ObjectRef.APIGroup,
			event.ObjectRef.APIVersion,
			event.ObjectRef.Resource,
			event.ObjectRef.Name,
			event.Stage,
			event.StageTimestamp.Time,
			event.Verb,
			event.ResponseStatus.Code,
			event.ResponseStatus.Status,
			event.ResponseStatus.Reason,
			event.ResponseStatus.Message)
		if err != nil {
			klog.Error(err)
		}
	}
}

func insertE(eventList *[]audit.Event) {
	db, err := sql.Open("mysql", "root:tmax@tcp(mysql-service.hypercloud4-system.svc:3306)/audit?parseTime=true")
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()

	for _, event := range *eventList {
		_, err := db.Exec(AUDIT_INSERT_QUERY,
			event.AuditID,
			event.User.Username,
			event.UserAgent,
			event.ObjectRef.Namespace,
			event.ObjectRef.APIGroup,
			event.ObjectRef.APIVersion,
			event.ObjectRef.Resource,
			event.ObjectRef.Name,
			event.Stage,
			event.StageTimestamp.Time,
			event.Verb,
			event.ResponseStatus.Code,
			event.ResponseStatus.Status,
			event.ResponseStatus.Reason,
			event.ResponseStatus.Message)
		if err != nil {
			klog.Error(err)
		}
	}
}

func get(query string) *audit.EventList {
	db, err := sql.Open("mysql", "root:tmax@tcp(mysql-service.hypercloud4-system.svc:3306)/audit?parseTime=true")
	if err != nil {
		klog.Error(err)
	}
	defer db.Close()
	rows, err := db.Query(query)
	if err != nil {
		rows.Close()
		klog.Error(err)
	}
	defer rows.Close()

	eventList := audit.EventList{}
	for rows.Next() {
		event := audit.Event{
			ObjectRef:      &audit.ObjectReference{},
			ResponseStatus: &metav1.Status{},
		}
		err := rows.Scan(
			&event.AuditID,
			&event.User.Username,
			&event.UserAgent,
			&event.ObjectRef.Namespace,
			&event.ObjectRef.APIGroup,
			&event.ObjectRef.APIVersion,
			&event.ObjectRef.Resource,
			&event.ObjectRef.Name,
			&event.Stage,
			&event.StageTimestamp.Time,
			&event.Verb,
			&event.ResponseStatus.Code,
			&event.ResponseStatus.Status,
			&event.ResponseStatus.Reason,
			&event.ResponseStatus.Message)
		if err != nil {
			rows.Close()
			klog.Error(err)
		}
		eventList.Items = append(eventList.Items, event)
	}

	return &eventList

}

package server

import (
	"encoding/json"
	"fmt"
	"github.com/creack/httpreq"
	"github.com/ryansxy/kevent/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"time"
)

var collectionName = "k8s1"

type EventChange struct {
	Reason    string  `json:"reason"`
	Message   string  `json:"message"`
	Count     int32   `json:"count"`
	FirstTime v1.Time `json:"first_time"`
	LastTime  v1.Time `json:"last_time"`
	Type      string  `json:"type"`
}
type EventsSample struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Changes   []EventChange
}

type EventResponseItem struct {
	ResourceName      string       `json:"resourceName"`
	ResourceType      string       `json:"resourceType"`
	ResourceNamespace string       `json:"resourceNamespace"`
	EventType         string       `json:"eventType"`
	EventSource       string       `json:"eventSource"`
	Message           string       `json:"message"`
	Timestamp         string       `json:"timestamp"`
	RawEvent          corev1.Event `json:"rawEvent"`
}

func FindEvents(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, "Paese http params is failed: "+err.Error(), 400)
		return
	}
	reqmessage := model.QFEvents{}
	if err := (httpreq.ParsingMap{
		{Field: "clustername", Fct: httpreq.ToString, Dest: &reqmessage.Clustername},
		{Field: "name", Fct: httpreq.ToString, Dest: &reqmessage.Name},
		{Field: "namespace", Fct: httpreq.ToString, Dest: &reqmessage.NameSpace},
		{Field: "kind", Fct: httpreq.ToString, Dest: &reqmessage.Kind},
	}.Parse(request.Form)); err != nil {
		http.Error(writer, "request params is invald", 400)
		return
	}
	if request.Body == nil {
		http.Error(writer, "Please send a request body", 400)
		return
	}

	if reqmessage.Clustername == "" {
		http.Error(writer, "the request message clustername must  is required", 400)
		return
	}

	events, err := model.FindEvents(reqmessage.Clustername, &reqmessage)
	if nil != err {
		http.Error(writer, fmt.Sprintf("the database response  error  ：%s", err.Error()), 500)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(events); err != nil {
		http.Error(writer, fmt.Sprintf("json format error  ：%s", err.Error()), 200)
		return
	}

}

// 获取 events 简版
func FindEventsSample(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, "Paese http params is failed: "+err.Error(), 400)
		return
	}
	reqmessage := model.QFEvents{}
	if err := (httpreq.ParsingMap{
		{Field: "clustername", Fct: httpreq.ToString, Dest: &reqmessage.Clustername},
		{Field: "name", Fct: httpreq.ToString, Dest: &reqmessage.Name},
		{Field: "namespace", Fct: httpreq.ToString, Dest: &reqmessage.NameSpace},
		{Field: "kind", Fct: httpreq.ToString, Dest: &reqmessage.Kind},
	}.Parse(request.Form)); err != nil {
		http.Error(writer, "request params is invald", 400)
		return
	}
	if request.Body == nil {
		http.Error(writer, "Please send a request body", 400)
		return
	}

	fmt.Println(reqmessage)
	if reqmessage.Clustername == "" {
		http.Error(writer, "the request message clustername must  is required", 400)
		return
	}

	events, err := model.FindEvents(reqmessage.Clustername, &reqmessage)
	if nil != err {
		http.Error(writer, fmt.Sprintf("the database response  error  ：%s", err.Error()), 500)
		return
	}
	es := EventsSample{}
	if events != nil && len(events) > 0 {
		es.Changes = []EventChange{}
		for _, v := range events {
			if v != nil {
				es.Kind = v.InvolvedObject.Kind
				es.Name = v.InvolvedObject.Name
				es.Namespace = v.InvolvedObject.Namespace
				es.Changes = append(es.Changes, EventChange{
					Reason:    v.Reason,
					Message:   v.Message,
					Type:      v.Type,
					Count:     v.Count,
					FirstTime: v.FirstTimestamp,
					LastTime:  v.LastTimestamp,
				})
			}
		}
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(es); err != nil {
		http.Error(writer, fmt.Sprintf("json format error  ：%s", err.Error()), 200)
		return
	}

}

// 根据 type 类型获取 指定的 event
func FindEventsByType(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, "Paese http params is failed: "+err.Error(), 400)
		return
	}
	reqmessage := model.QFEvents{}
	if err := (httpreq.ParsingMap{
		{Field: "type", Fct: httpreq.ToString, Dest: &reqmessage.Type},
	}.Parse(request.Form)); err != nil {
		http.Error(writer, "request params is invald", 400)
		return
	}
	if request.Body == nil {
		http.Error(writer, "Please send a request body", 400)
		return
	}

	fmt.Println(reqmessage)

	events, err := model.FindEvents(collectionName, &reqmessage)
	if nil != err {
		http.Error(writer, fmt.Sprintf("the database response  error  ：%s", err.Error()), 500)
		return
	}

	var es []EventResponseItem
	if events != nil && len(events) > 0 {
		for _, v := range events {
			if v != nil {
				eventResponseItem := translateEvent2EventResponse(v)
				es = append(es, *eventResponseItem)
			}
		}
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(es); err != nil {
		http.Error(writer, fmt.Sprintf("json format error  ：%s", err.Error()), 200)
		return
	}
}

func translateEvent2EventResponse(event *corev1.Event) *EventResponseItem {
	var resourceName, resourceType, resourceNamespace string

	resourceType = event.InvolvedObject.Kind
	resourceName = event.InvolvedObject.Name
	resourceNamespace = event.InvolvedObject.Namespace
	source := fmt.Sprintf("Event come from Component(%s)", event.Source.Component)
	if event.Source.Host != "" {
		source = fmt.Sprintf("Event come from Component(%s) Host(%s)", event.Source.Component, event.Source.Host)
	}
	var cstZone = time.FixedZone("CST", 8*3600)
	return &EventResponseItem{
		ResourceName:      resourceName,
		ResourceType:      resourceType,
		ResourceNamespace: resourceNamespace,
		EventType:         event.Type,
		EventSource:       source,
		Message:           event.Message,
		Timestamp:         event.FirstTimestamp.In(cstZone).Format("2006-01-02 15:04:05"),
		RawEvent:          *event,
	}
}

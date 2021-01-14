package model

import (
	"github.com/ryansxy/kevent/store"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"upper.io/db.v3"
)

type QFEvents struct {
	Name        string
	NameSpace   string
	Kind        string
	Clustername string
	Type        string
}

func FindEvents(colletionname string, qr *QFEvents) (events []*v1.Event, err error) {
	condmap := make(map[string]db.Comparison, 10)
	if qr.Kind != "" {
		condmap["involvedobject.kind"] = db.Eq(qr.Kind)
	}
	if qr.Name != "" {
		condmap["involvedobject.name"] = db.Eq(qr.Name)
	}
	if qr.NameSpace != "" {
		condmap["involvedobject.namespace"] = db.Eq(qr.NameSpace)
	}
	if qr.Type != "" {
		condmap["type"] = db.Eq(qr.Type)
	}
	conds := make(map[interface{}]interface{})
	for k, v := range condmap {
		conds[k] = v
	}
	events = []*v1.Event{}
	if _, e := store.DefaultMongoStore.Find(colletionname, conds, &events); e != nil {
		err = e
		return
	} else {
	}

	return
}

func DeleteEvents(colletionname string, dt metav1.Time) (err error) {
	conds := make(map[interface{}]interface{})

	conds["objectmeta.creationtimestamp"] = db.Lte(dt)
	for {
		store.DefaultMongoStore.Delete(colletionname, conds)
		//time.Sleep(time.Second * 30)
	}
	return nil
}

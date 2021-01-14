package k8s

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/ryansxy/kevent/store"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// this const var is prometheus metrics
var (
	kubernetesWarningEventCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kubernetes_eventrouter_warnings_total",
		Help: "Total number of warning events in the kubernetes cluster",
	}, []string{
		"involved_object_kind",
		"involved_object_name",
		"involved_object_namespace",
		"reason",
		"source",
		"clustername",
	})
	kubernetesNormalEventCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kubernetes_eventrouter_normal_total",
		Help: "Total number of normal events in the kubernetes cluster",
	}, []string{
		"involved_object_kind",
		"involved_object_name",
		"involved_object_namespace",
		"reason",
		"source",
		"clustername",
	})
)

// prometheus register the metrics
func init() {
	prometheus.MustRegister(kubernetesWarningEventCounterVec)
	prometheus.MustRegister(kubernetesNormalEventCounterVec)
}

// EventRouter is responsible for maintaining a stream of kubernetes
// system Events and pushing them to another channel for storage
type EventRouter struct {
	kubeClient kubernetes.Interface

	// store of events populated by the shared informer
	eLister corelisters.EventLister

	// returns true if the event store has been synced
	eListerSynched cache.InformerSynced

	// cluster name
	Clustername string
}

// NewEventRouter will create a new event router using the input params
func NewEventRouter(kubeClient kubernetes.Interface, eventsInformer coreinformers.EventInformer, clustername string) *EventRouter {
	er := &EventRouter{
		kubeClient: kubeClient,
	}
	// give  event  informer the  eventHandler
	eventsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    er.addEvent,
		UpdateFunc: er.updateEvent,
		DeleteFunc: er.deleteEvent,
	})
	er.eLister = eventsInformer.Lister()
	er.eListerSynched = eventsInformer.Informer().HasSynced
	er.Clustername = clustername
	return er
}

// Run starts the EventRouter/Controller.
func (er *EventRouter) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer logrus.Infof("Shutting down EventRouter")

	logrus.Infof("Starting EventRouter")

	// here is where we kick the caches into gear
	if !cache.WaitForCacheSync(stopCh, er.eListerSynched) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	<-stopCh
}

// addEvent is called when an event is created, or during the initial list
func (er *EventRouter) addEvent(obj interface{}) {
	e := obj.(*v1.Event)
	prometheusEvent(e, er)

	// e returns true before t
	// 如果时间的创建时间再当前时间之前
	if (e.CreationTimestamp).Before(&nowtime) {
		return
	}

	err := store.DefaultMongoStore.Insert(er.Clustername, e)
	if nil != err {
		log.Warn("insert event %s error %s", e.Name, err.Error())
	}
}

// updateEvent is called any time there is an update to an existing event
func (er *EventRouter) updateEvent(objOld interface{}, objNew interface{}) {
	eOld := objOld.(*v1.Event)
	eNew := objNew.(*v1.Event)
	prometheusEvent(eNew, er)
	if eOld.ResourceVersion != eNew.ResourceVersion {
		err := store.DefaultMongoStore.Insert(er.Clustername, eNew)
		if nil != err {
			log.Warn("insert event %s error %s", eNew.Name, err.Error())
		}
	}
}

// prometheusEvent is called when an event is added or updated
func prometheusEvent(event *v1.Event, er *EventRouter) {
	var counter prometheus.Counter
	var err error

	if event.Type == "Normal" {
		counter, err = kubernetesNormalEventCounterVec.GetMetricWithLabelValues(
			event.InvolvedObject.Kind,
			event.InvolvedObject.Name,
			event.InvolvedObject.Namespace,
			event.Reason,
			event.Source.Host,
			er.Clustername,
		)
	} else if event.Type == "Warning" {
		counter, err = kubernetesWarningEventCounterVec.GetMetricWithLabelValues(
			event.InvolvedObject.Kind,
			event.InvolvedObject.Name,
			event.InvolvedObject.Namespace,
			event.Reason,
			event.Source.Host,
			er.Clustername,
		)
	}

	if err != nil {
		logrus.Warnf("prometheus event error: " + err.Error())
	} else {
		counter.Add(1)
	}
}

func (er *EventRouter) deleteEvent(obj interface{}) {
	/*
	e := obj.(*v1.Event)
	e = e*/
	//logrus.Infof("Event Deleted from the system:\n%v", e)
}

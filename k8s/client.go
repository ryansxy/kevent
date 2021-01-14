package k8s

import (
	"context"
	"github.com/ryansxy/kevent/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sync"
	"time"
)

type ClusterMessage struct {
	ClusterName string
	Kubeconfig  string
}

var clusterNs = make([]*ClusterMessage, 0)
var nowtime v1.Time
//var dt v1.Time

func init() {
	nowt := time.Now()
	nowtime = v1.NewTime(nowt)
	//  deletetime:=nowt.Add(-time.Minute*2)
	// dt = v1.NewTime(deletetime)
}

// createClient will create kubernetes client
func createClient(clustername, kubeconfig string) (string, kubernetes.Interface) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		logrus.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	// creates the clientset from kubeconfig
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	return clustername, clientset
}

func StartEventsCollection(ctx context.Context, clusters []*ClusterMessage) {
	wg := sync.WaitGroup{}
	stopCh := ctx.Done()
	for _, cm := range clusters {
		clustername, clientset := createClient(cm.ClusterName, cm.Kubeconfig)
		sharedInformers := informers.NewSharedInformerFactory(clientset, viper.GetDuration("resync-interval"))
		eventsInformer := sharedInformers.Core().V1().Events()

		eventRouter := NewEventRouter(clientset, eventsInformer, clustername)

		// Startup the EventRouter
		wg.Add(1)
		go func() {
			defer wg.Done()
			eventRouter.Run(stopCh)
		}()
		// Startup the Informer(s)
		logrus.Infof("Starting shared Informer(s)")
		sharedInformers.Start(stopCh)
		// 定期清理events
		go func() {
			ticker := time.NewTicker(time.Minute * 30)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					var dt v1.Time
					nowt := time.Now()
					nowtime = v1.NewTime(nowt)
					deletetime := nowt.Add(-time.Hour * 1)
					dt = v1.NewTime(deletetime)
					model.DeleteEvents(cm.ClusterName, dt)
				}
			}
		}()

	}
	wg.Wait()
}

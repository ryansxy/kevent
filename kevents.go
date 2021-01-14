package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ryansxy/kevent/k8s"
	"github.com/ryansxy/kevent/server"
	"github.com/ryansxy/kevent/store"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"strings"
	"sync"
)

func kevents(c *cli.Context) {
	var err error
	var clusters []*k8s.ClusterMessage
	// init kube-config
	if clusters, err = processClusterMessage(c); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err = initDb(c); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	stopC := sigHandler()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-stopC
		cancel()
	}()
	if nil != clusters && len(clusters) > 0 {
		go func() {
			k8s.StartEventsCollection(ctx, clusters)
			wg.Done()
		}()
		go func() {
			sc := &server.HTTPServer{
				Addr: c.String("listen-address"),
			}
			err = server.ServeHTTP(ctx, sc)
			wg.Done()
		}()
	}
	wg.Wait()
	logrus.Warnf("Exiting main()")
	os.Exit(1)
}

func processClusterMessage(c *cli.Context) (cms []*k8s.ClusterMessage, err error) {
	// 校验 kubeconfig 合法性
	kp := c.StringSlice("kubeconfig")
	for _, str := range kp {
		kubeconfigs := strings.Split(str, "-")
		if len(kubeconfigs) != 2 {
			err = errors.New("The parameter conf format is incorrect.")
			return
		}
		if cms == nil {
			cms = []*k8s.ClusterMessage{}
		}
		cm := &k8s.ClusterMessage{kubeconfigs[0], kubeconfigs[1]}
		cms = append(cms, cm)
	}
	return
}

func initDb(c *cli.Context) (err error) {
	config := store.MongoConfig{}
	if c.String("mongo-address") != "" {
		config.Host = c.String("mongo-address")
	}
	if c.String("mongo-db") != "" {
		config.Database = c.String("mongo-db")
	}
	if c.String("mongo-user") != "" {
		config.User = c.String("mongo-user")
	}
	if c.String("mongo-passwd") != "" {
		config.Password = c.String("mongo-passwd")
	}
	err = store.InitMongodbStore(&config)
	return
}

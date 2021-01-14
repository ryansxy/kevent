package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"syscall"
)

// setup a signal hander to gracefully exit
func sigHandler() <-chan struct{} {
	stop := make(chan struct{})
	go func() {
		// set up a operating system signal. The usual underlying implementation is operating system-dependent.  on Unix it is syscall.Signal.
		c := make(chan os.Signal, 1)
		// Notify causes package signal to relay incoming signals to c.
		signal.Notify(c,
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // Termination Request
			syscall.SIGSEGV, // FullDerp
			syscall.SIGABRT, // Abnormal termination
			syscall.SIGILL,  // illegal instruction
			syscall.SIGFPE) // floating point - this is why we can't have nice things
		sig := <-c
		logrus.Warnf("Signal (%v) Detected, Shutting Down", sig)
		close(stop)
	}()
	return stop
}

var flags = []cli.Flag{
	// base config
	cli.StringFlag{
		EnvVar: "HTTP_LISTEN_ADDRESS",
		Name:   "listen-address",
		Usage:  "http server 服务地址",
		Value:  ":8080",
	},
	cli.StringSliceFlag{
		EnvVar: "KUBE_CONFIG_PATH",
		Name:   "kubeconfig",
		Usage:  "kubeconfig 路径，支持多次使用（多K8S集群）",
		Value:  &cli.StringSlice{},
	},
	cli.StringFlag{
		EnvVar: "MONGODB_HOST",
		Name:   "mongo-address",
		Usage:  "mongodb 服务地址（含端口号）",
		Value:  "",
	},
	cli.StringFlag{
		EnvVar: "MONGODB_DB",
		Name:   "mongo-db",
		Usage:  "mongodb 数据库名",
		Value:  "events",
	},
	cli.StringFlag{
		EnvVar: "MONGODB_USER",
		Name:   "mongo-user",
		Usage:  "mongodb 服务登录用户",
		Value:  "",
	},
	cli.StringFlag{
		EnvVar: "MONGODB_PASSWD",
		Name:   "mongo-passwd",
		Usage:  "mongodb 服务密码",
		Value:  "",
	},
}

func before(c *cli.Context) error { return nil }

func main() {
	app := cli.NewApp()
	app.Name = "kevents"
	app.Version = ""
	app.Usage = "kevents"
	app.Action = kevents
	app.Flags = flags
	app.Before = before
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

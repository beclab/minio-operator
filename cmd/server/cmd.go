package server

import (
	"context"
	"os"

	"bytetrade.io/web3os/minio-operator/cmd/base"
	"k8s.io/klog/v2"
)

type ServerCommand struct {
	*base.BaseCommand
	manager *Manager
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewServer() *ServerCommand {
	getEnvArgs(&base.ETCD_CAFILE, "ETCD_CAFILE")
	getEnvArgs(&base.ETCD_CERTFILE, "ETCD_CERTFILE")
	getEnvArgs(&base.ETCD_KEYFILE, "ETCD_KEYFILE")
	getEnvArgs(&base.ETCD_SERVER, "ETCD_SERVER")
	getEnvArgs(&base.MINIO_NODE_VOLUME, "MINIO_NODE_VOLUME")
	getEnvArgs(&NODE, "NODE")

	baseCmd := base.New()
	ctx, cancel := context.WithCancel(context.Background())
	return &ServerCommand{
		BaseCommand: baseCmd,
		manager:     &Manager{Client: baseCmd.Client},
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *ServerCommand) Exec() error {
	watcher := Watcher{
		Client:  s.Client,
		Manager: s.manager,
		Ctx:     s.ctx,
		Server:  s,
	}

	klog.Info("start minio operator, try to start minio")
	config, err := s.GetClusterConfig(s.ctx)
	if err != nil {
		klog.Error("get minio cluster config error, ", err)
		return err
	}

	err = s.manager.ReloadMinIOService(config)
	if err != nil {
		klog.Error("start minio service error, ", err)
		return err
	}

	SetupSignalHandler(s.ctx, s.cancel)

	return watcher.Start()
}

func getEnvArgs(arg *string, env string) {
	v := os.Getenv(env)
	if v != "" {
		*arg = v
	}
}

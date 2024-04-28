package server

import (
	"context"

	"bytetrade.io/web3os/minio-operator/pkg/minio"
	"bytetrade.io/web3os/minio-operator/pkg/storage/client"
	"bytetrade.io/web3os/minio-operator/pkg/storage/model"
	"github.com/coreos/etcd/storage/storagepb"
	"k8s.io/klog/v2"
)

type Watcher struct {
	Client  client.Client
	Ctx     context.Context
	Manager *Manager
	Server  *ServerCommand
}

func (w *Watcher) Start() error {
	for {
		watchChan := w.Client.WatchList(w.Ctx, model.KEY_PREFIX)
		select {
		case <-w.Ctx.Done():
			klog.Info("cluster watcher stopped")
			return nil
		case resp, ok := <-watchChan:
			if !ok || resp.Err() != nil {
				if resp.Err() != nil {
					klog.Error("cluster watch chan response error, ", resp.Err())
				}
				klog.Info("cluster watch chan closed, retry watch")
				continue
			}

			var prevConfig *minio.ClusterConfig = nil
			for _, event := range resp.Events {
				switch event.Type {
				case storagepb.PUT:
					klog.Info("cluster config updated, ", event.Kv, ", ", string(event.Kv.Value))
					config, err := w.Server.GetClusterConfig(w.Ctx)
					if err != nil {
						klog.Error("get cluster config error, ", err)
						continue
					}

					if prevConfig != nil {
						if len(prevConfig.Nodes) == len(config.Nodes) && prevConfig.Mode == config.Mode {
							if prevConfig.Mode == minio.SNMD &&
								prevConfig.Nodes[0].Drivers.Num == config.Nodes[0].Drivers.Num {
								klog.Info("cluster config does not be changed, ignore the event")
								continue
							}
						}
					}

					prevConfig = config

					err = w.Manager.ReloadMinIOService(config)
					if err != nil {
						klog.Error("reload minio service error, ", err)
						return err
					}
				case storagepb.DELETE:
					klog.Warning("node removed from cluster, ", event.Kv, ", ", string(event.Kv.Value))
				}
			}
		}
	}
}

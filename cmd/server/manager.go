package server

import (
	"errors"
	"os"

	"bytetrade.io/web3os/minio-operator/cmd/base"
	"bytetrade.io/web3os/minio-operator/pkg/minio"
	"bytetrade.io/web3os/minio-operator/pkg/storage/client"
	"bytetrade.io/web3os/minio-operator/pkg/storage/model"
	"k8s.io/klog/v2"
)

type Manager struct {
	Client client.Client
}

func (m *Manager) createOrUpdateMinIOService() error {
	err := os.WriteFile(minio.MinIOServiceFile, []byte(minio.MinIOService), 0644)
	if err != nil {
		klog.Error("create service in ", minio.MinIOServiceFile, ", error, ", err)
		return err
	}

	return minio.ReloadMinIOService()

}

func (m *Manager) createMinIOEnv(config *minio.ClusterConfig) error {
	return config.CreateDefaultEnv()
}

func (m *Manager) ReloadMinIOService(config *minio.ClusterConfig) error {
	_, err := os.Stat(minio.DefaultConfigFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		klog.Error("get file info error, ", err, ", ", minio.DefaultConfigFile)
		return err
	}

	if func() bool {
		for _, n := range config.Nodes {
			if model.GetNodeKey(n.Hostname) == NODE {
				return false
			}
		}

		return true
	}() {
		klog.Warning("current node is not one of cluster nodes")
		return nil
	}

	// node is running
	if err == nil {
		mode, err := minio.GetCurrentMode()
		if err != nil {
			return err
		}

		// ONLY SNMD mode or going to be changed to SNMD mode,
		// cluster sys file needs to be cleared to reload the cluster
		if config.Mode == minio.SNMD || mode == minio.SNMD {
			klog.Infof("cluster mode will be changed from %s to %s, clear current sys config", mode, config.Mode)
			err = minio.ClearMinIOSysfile(base.MINIO_NODE_VOLUME)
			if err != nil {
				return err
			}
		}
	}

	klog.Info("creating minio default env config")
	err = m.createMinIOEnv(config)
	if err != nil {
		return err
	}

	klog.Info("creating or updating minio service, and reload service")
	return m.createOrUpdateMinIOService()
}

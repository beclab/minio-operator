package base

import (
	"context"
	"errors"
	"os"

	"bytetrade.io/web3os/minio-operator/pkg/minio"
	"bytetrade.io/web3os/minio-operator/pkg/storage/client"
	"bytetrade.io/web3os/minio-operator/pkg/storage/model"
	"bytetrade.io/web3os/minio-operator/pkg/utils"
	"k8s.io/klog/v2"
)

type BaseCommand struct {
	Client client.Client
}

func New() *BaseCommand {
	config := client.ETCDConfig{
		Endpoints: []string{ETCD_SERVER},
		TLSConfig: client.Config{
			CAFile:   ETCD_CAFILE,
			CertFile: ETCD_CERTFILE,
			KeyFile:  ETCD_KEYFILE,
		},
	}

	client, err := client.New(config)
	if err != nil {
		klog.Error("create etcd client error")
		panic(err)
	}

	return &BaseCommand{Client: client}
}

func (m *BaseCommand) GetClusterConfig(ctx context.Context) (*minio.ClusterConfig, error) {
	nodes, err := m.Client.List(ctx, model.KEY_PREFIX)
	if err != nil && err != client.ErrNotFound {
		return nil, err
	}

	clusterData, err := m.Client.Get(ctx, model.CLUSTER_KEY)
	if err != nil && err != client.ErrNotFound {
		return nil, err
	}

	var cluster model.Cluster
	v := model.DataValue(clusterData.Data)
	err = v.UnmarshalTo(&cluster)
	if err != nil {
		klog.Error("parse cluster data error, ", err, ", ", string(v))
		return nil, err
	}

	config := minio.ClusterConfig{
		RootUser:     cluster.RootUser,
		RootPassword: cluster.RootPassword,
	}

	switch len(nodes) {
	case 0:
		klog.Error("cluster not found")
		return nil, errors.New("cluster not found")
	case 1:
		data := model.DataValue(nodes[0].Data)
		var node model.Node
		err = data.UnmarshalTo(&node)
		if err != nil {
			klog.Error("parse node data error, ", err, ", ", string(data))
			return nil, err
		}

		switch {
		case node.Drivers.Num > 1:
			config.Mode = minio.SNMD
		default:
			config.Mode = minio.SNSD
		}

		config.Nodes = []*model.Node{&node}
	default:
		config.Mode = minio.MNMD
		for _, n := range nodes {
			data := model.DataValue(n.Data)
			var node model.Node
			err = data.UnmarshalTo(&node)
			if err != nil {
				klog.Error("parse node data error, ", err, ", ", string(data))
				return nil, err
			}

			config.Nodes = append(config.Nodes, &node)
		}
	}

	return &config, nil
}

func (c *BaseCommand) CreateOperatorServerEnv(nodeKey string) error {
	env := map[string]string{
		"ETCD_CAFILE":   ETCD_CAFILE,
		"ETCD_CERTFILE": ETCD_CERTFILE,
		"ETCD_KEYFILE":  ETCD_KEYFILE,
		"ETCD_SERVER":   ETCD_SERVER,

		"MINIO_NODE_VOLUME": MINIO_NODE_VOLUME,
		"NODE":              nodeKey,
	}

	klog.Info("creating minio-operator server env")
	return utils.WritePropertiesFile(env, minio.DefaultOperatorEnvFile)
}

func (c *BaseCommand) CreateOperatorService() error {
	err := os.WriteFile(minio.OperatorServiceFile, []byte(minio.MinIOOperatorService), 0644)
	if err != nil {
		klog.Error("create service in ", minio.OperatorServiceFile, ", error, ", err)
		return err
	}

	klog.Info("creating minio-operator service")
	return minio.ReloadMinIOOperatorService()
}

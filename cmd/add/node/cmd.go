package node

import (
	"context"
	"errors"
	"os"
	"strings"

	"bytetrade.io/web3os/minio-operator/cmd/base"
	"bytetrade.io/web3os/minio-operator/pkg/minio"
	"bytetrade.io/web3os/minio-operator/pkg/storage/client"
	"bytetrade.io/web3os/minio-operator/pkg/storage/model"
	"k8s.io/klog/v2"
)

type NodeCommand struct {
	*base.BaseCommand
}

func NewNode() *NodeCommand {
	return &NodeCommand{BaseCommand: base.New()}
}

func (c *NodeCommand) Exec() error {
	ctx := context.TODO()
	defer c.BaseCommand.Client.Close()

	var err error
	if err = c.validateConfig(); err != nil {
		return err
	}

	var nodeNum = 0
	if nodeNum, err = c.validateCluster(ctx); err != nil {
		return err
	}

	nodeKey, err := c.updateClusterConfig(ctx, nodeNum+1)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			c.clearClusterConfig(ctx, nodeKey)
		}
	}()

	if err = c.CreateOperatorServerEnv(nodeKey); err != nil {
		return err
	}

	if err = c.CreateOperatorService(); err != nil {
		return err
	}

	return nil
}

func (c *NodeCommand) validateCluster(ctx context.Context) (int, error) {
	// check driver path
	if !strings.HasPrefix(base.MINIO_NODE_VOLUME, "/") {
		klog.Error("invalid volume path, ", base.MINIO_NODE_VOLUME)
		return 0, errors.New("invalid path")
	}

	volumes, err := minio.VolumeExpend(base.MINIO_NODE_VOLUME)
	if err != nil {
		klog.Error("can not expand node volume, ", err, ", ", base.MINIO_NODE_VOLUME)
		return 0, err
	}

	if len(volumes) <= 1 {
		klog.Error("`add node` must be multi-driver mode")
		return 0, errors.New("unsupported mode")
	}

	if m, err := minio.GetClusterMode(volumes); err != nil || m == minio.MNMD {
		if err != nil {
			return 0, err
		}

		klog.Error("`add node` must add node one by one")
		return 0, errors.New("unsupported mode")
	}

	for _, v := range volumes {
		info, err := os.Lstat(v[0])
		if err != nil {
			klog.Error("get path stat error, ", err, ", ", v)
			return 0, err
		}

		// dir owner should be minio
		if !info.IsDir() {
			// whether the file is soft link to anther path or not
			if info.Mode()&os.ModeSymlink != 0 {
				info, err = os.Lstat(v[0])
				if err != nil {
					klog.Error("get path stat error, ", err, ", ", v)
					return 0, err
				}

				if !info.IsDir() {
					klog.Error("volume is not a dir, ", v)
					return 0, errors.New("invalid path")
				}
			} else {
				klog.Error("volume is not a dir, ", v)
				return 0, errors.New("invalid path")
			}
		}
	}

	// check the cluster nodes config
	nodes, err := c.BaseCommand.Client.List(ctx, model.KEY_PREFIX)
	if err != nil && err != client.ErrNotFound {
		return 0, err
	}

	nodeNum := len(nodes)
	switch {
	case nodeNum == 0:
		return 0, errors.New("cluster is not found")

	case nodeNum == 1:
		klog.Info("current cluser is in single-node mode")

		var node model.Node
		data := model.DataValue(nodes[0].Data)
		err = data.UnmarshalTo(&node)
		if err != nil {
			klog.Error("parse node error, ", err, ", ", string(data))
			return 0, err
		}

		if node.Drivers.Num == 1 {
			klog.Error("to add the second node, the current cluster must have the SNMD mode")
			return 0, errors.New("invalid volume")
		}

	default:
		klog.Info("current cluser is MNMD mode")

	}

	klog.Info("new node will be added to cluster")
	return nodeNum, nil
}

func (c *NodeCommand) validateConfig() error {
	if MINIO_NODE_ADDRESS == "" {
		return errors.New("node ip address invalid")
	}

	if base.MINIO_NODE_VOLUME == "" {
		return errors.New("node data volume invalid")
	}

	return nil
}

func (c *NodeCommand) updateClusterConfig(ctx context.Context, index int) (string, error) {
	path := base.MINIO_NODE_VOLUME

	expanded, err := minio.VolumeExpend(path)
	if err != nil {
		klog.Error("invalid path, ", err, ", ", path)
		return "", err
	}

	// path is validated
	node := model.Node{
		IP:       MINIO_NODE_ADDRESS,
		Hostname: minio.NodeName(index),
		Port:     "9000",

		Drivers: model.Drivers{
			Path: base.MINIO_NODE_VOLUME,
			Num:  len(expanded),
		},
	}

	data := model.DataValue{}

	if err := data.MarshalFrom(&node); err != nil {
		klog.Error("cannot parse node config, ", err, ", ", node)
		return "", err
	}

	key := model.GetNodeKey(node.Hostname)

	if err := c.BaseCommand.Client.Create(ctx, model.GetNodeKey(node.Hostname), data); err != nil {
		klog.Error("save node config to storage error, ", err, ", ", key)
		return "", nil
	}

	return key, nil
}

func (c *NodeCommand) clearClusterConfig(ctx context.Context, key string) error {
	node, err := c.BaseCommand.Client.Get(ctx, key)
	if err != nil {
		if err == client.ErrNotFound {
			return nil
		}

		klog.Error("find node config error, ", err, ", ", key)
		return err
	}

	err = c.BaseCommand.Client.Delete(ctx, key, node.Modified)
	if err != nil {
		klog.Error("clear node config error, ", err, ", ", key)
	}

	return err
}

package init_cmd

import (
	"context"
	"errors"
	"os"
	"strings"

	"bytetrade.io/web3os/minio-operator/cmd/base"
	"bytetrade.io/web3os/minio-operator/pkg/minio"
	"bytetrade.io/web3os/minio-operator/pkg/storage/client"
	"bytetrade.io/web3os/minio-operator/pkg/storage/model"
	"bytetrade.io/web3os/minio-operator/pkg/utils"
	"k8s.io/klog/v2"
)

type InitCommand struct {
	*base.BaseCommand
}

func NewInit() *InitCommand {
	return &InitCommand{BaseCommand: base.New()}
}

func (c *InitCommand) Exec() error {
	ctx := context.TODO()
	defer c.BaseCommand.Client.Close()

	var err error
	if err = c.validateConfig(); err != nil {
		return err
	}

	if err = c.validateCluster(ctx); err != nil {
		return err
	}

	nodeKey, err := c.createClusterConfig(ctx)
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

func (c *InitCommand) validateCluster(ctx context.Context) error {
	nodes, err := c.BaseCommand.Client.List(ctx, model.KEY_PREFIX)
	if err != nil && err != client.ErrNotFound {
		return err
	}

	if len(nodes) > 0 {
		return errors.New("there is a cluster running, do not init twice")
	}

	// check driver path
	if !strings.HasPrefix(base.MINIO_NODE_VOLUME, "/") {
		klog.Error("invalid volume path, ", base.MINIO_NODE_VOLUME)
		return errors.New("invalid path")
	}

	volumes, err := minio.VolumeExpend(base.MINIO_NODE_VOLUME)
	if err != nil {
		klog.Error("can not expand node volume, ", err, ", ", base.MINIO_NODE_VOLUME)
		return err
	}

	if m, err := minio.GetClusterMode(volumes); err != nil || m == minio.MNMD {
		if err != nil {
			return err
		}

		klog.Error("minio-operator init does not support multi nodes mode")
		return errors.New("unsupported mode")
	}

	for _, v := range volumes {

		info, err := os.Lstat(v[0])
		if err != nil {
			klog.Error("get path stat error, ", err, ", ", v)
			return err
		}

		// dir owner should be minio
		if !info.IsDir() {
			// whether the file is soft link to anther path or not
			if info.Mode()&os.ModeSymlink != 0 {
				info, err = os.Lstat(v[0])
				if err != nil {
					klog.Error("get path stat error, ", err, ", ", v)
					return err
				}

				if !info.IsDir() {
					klog.Error("volume is not a dir, ", v)
					return errors.New("invalid path")
				}
			} else {
				klog.Error("volume is not a dir, ", v)
				return errors.New("invalid path")
			}
		}
	}

	klog.Info("a new cluster will to be created")
	return nil
}

func (c *InitCommand) validateConfig() error {
	if MINIO_NODE_ADDRESS == "" {
		return errors.New("node ip address invalid")
	}

	if base.MINIO_NODE_VOLUME == "" {
		return errors.New("node data volume invalid")
	}

	return nil
}

func (c *InitCommand) createClusterConfig(ctx context.Context) (string, error) {
	path := base.MINIO_NODE_VOLUME

	expanded, err := minio.VolumeExpend(path)
	if err != nil {
		klog.Error("invalid path, ", err, ", ", path)
		return "", err
	}

	// path is validated, assume it's single node mod
	node := model.Node{
		IP:       MINIO_NODE_ADDRESS,
		Hostname: minio.NodeName(1),
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

	// create cluster info
	password := ROOT_PASSWORD
	if password == "" {
		password = utils.RandomString(16)
	}

	cluster := model.Cluster{
		RootUser:     minio.DefaultRootUser,
		RootPassword: password,
	}

	if err := data.MarshalFrom(&cluster); err != nil {
		klog.Error("cannot parse cluster config, ", err, ", ", cluster)
		return "", err
	}

	if err := c.BaseCommand.Client.Create(ctx, model.CLUSTER_KEY, data); err != nil {
		klog.Error("save cluster config to storage error, ", err, ", ", key)
		return "", nil
	}

	return key, nil
}

func (c *InitCommand) clearClusterConfig(ctx context.Context, key string) error {
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

	cluster, err := c.BaseCommand.Client.Get(ctx, model.CLUSTER_KEY)
	if err != nil {
		if err == client.ErrNotFound {
			return nil
		}

		klog.Error("find cluster config error, ", err, ", ", model.CLUSTER_KEY)
		return err
	}

	err = c.BaseCommand.Client.Delete(ctx, model.CLUSTER_KEY, cluster.Modified)
	if err != nil {
		klog.Error("clear clsuter config error, ", err, ", ", model.CLUSTER_KEY)
	}

	return err
}

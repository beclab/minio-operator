package driver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"bytetrade.io/web3os/minio-operator/cmd/base"
	"bytetrade.io/web3os/minio-operator/pkg/minio"
	"bytetrade.io/web3os/minio-operator/pkg/storage/client"
	"bytetrade.io/web3os/minio-operator/pkg/storage/model"
	"github.com/minio/pkg/v2/ellipses"
	"k8s.io/klog/v2"
)

type DriverCommand struct {
	*base.BaseCommand
}

func NewDriver() *DriverCommand {
	return &DriverCommand{BaseCommand: base.New()}
}

func (c *DriverCommand) Exec() error {
	ctx := context.TODO()
	defer c.BaseCommand.Client.Close()

	var err error
	if err = c.validateConfig(); err != nil {
		return err
	}

	if err = c.validateCluster(ctx); err != nil {
		return err
	}

	if err = c.updateClusterConfig(ctx); err != nil {
		return err
	}
	return nil
}

func (c *DriverCommand) validateCluster(ctx context.Context) error {
	nodes, err := c.BaseCommand.Client.List(ctx, model.KEY_PREFIX)
	if err != nil && err != client.ErrNotFound {
		return err
	}

	if len(nodes) == 0 {
		return errors.New("cluster is not found")
	}

	if len(nodes) > 1 {
		return errors.New("cluster is running on MNMD mode, cannot add drivers to it")
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

		klog.Error("`add driver` does not support multi nodes mode")
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

	klog.Info("new drivers will be added to node")
	return nil

}

func (c *DriverCommand) validateConfig() error {
	if base.MINIO_NODE_VOLUME == "" {
		return errors.New("node data volume invalid")
	}

	return nil
}

func (c *DriverCommand) updateClusterConfig(ctx context.Context) error {
	// add driver must be node minio-1
	key := model.GetNodeKey("minio-1")
	data, err := c.BaseCommand.Client.Get(ctx, key)
	if err != nil {
		klog.Error("get node config error, ", err, ", ", key)
		return err
	}

	nodeData := model.DataValue(data.Data)
	var node model.Node

	err = nodeData.UnmarshalTo(&node)
	if err != nil {
		klog.Error("parse node data error, ", err)
		return err
	}

	mode, err := minio.GetCurrentMode()
	if err != nil {
		return err
	}

	switch mode {
	case minio.SNSD:
		if ellipses.HasEllipses(base.MINIO_NODE_VOLUME) {
			// add more than one volumes
			volumes, err := minio.VolumeExpend(base.MINIO_NODE_VOLUME)
			if err != nil {
				klog.Error("can not expand node volume, ", err, ", ", base.MINIO_NODE_VOLUME)
				return err
			}

			if m, err := minio.GetClusterMode(volumes); err != nil || m == minio.MNMD {
				if err != nil {
					return err
				}

				klog.Error("minio-operator add driver does not support multi nodes mode")
				return errors.New("unsupported mode")
			}

			if start, pattern, ok := validateFirstAddingVolume(node.Drivers.Path, volumes[0][0]); !ok {
				return errors.New("invalid volume to be added")
			} else {
				node.Drivers.Num += len(volumes)
				node.Drivers.Path = fmt.Sprintf(pattern, fmt.Sprintf("{%d...%d}", start, start+len(volumes)))
			}

		} else {
			// check whether the adding volume index equals the previous volume index plus one or not
			if start, pattern, ok := validateFirstAddingVolume(node.Drivers.Path, base.MINIO_NODE_VOLUME); !ok {
				return errors.New("invalid volume to be added")
			} else {
				node.Drivers.Num = 2
				node.Drivers.Path = fmt.Sprintf(pattern, fmt.Sprintf("{%d...%d}", start, start+1))
			}

		}
	case minio.SNMD:
		currentVolumesStart, currentVolumesEnd, err := minio.FindVolumeIndex(node.Drivers.Path)
		if err != nil {
			klog.Error("can not expand current node volume, ", err, ", ", base.MINIO_NODE_VOLUME)
			return err
		}

		var (
			volumeEnd    int
			replceVolume string
		)
		// adding volume is multi-driver mode
		if ellipses.HasEllipses(base.MINIO_NODE_VOLUME) {
			volumesStart, end, err := minio.FindVolumeIndex(base.MINIO_NODE_VOLUME)
			if err != nil {
				klog.Error("can not expand add driver volume, ", err, ", ", base.MINIO_NODE_VOLUME)
				return err
			}
			if currentVolumesEnd != volumesStart-1 {
				klog.Error("there is a gap from current volume and added volume")
				return errors.New("invalid volume")
			}

			volumeEnd = end
			replceVolume, err = minio.ReplaceVolumeIndex(base.MINIO_NODE_VOLUME, currentVolumesStart, volumeEnd)
			if err != nil {
				return err
			}
		} else {
			volumes, err := minio.VolumeExpend(node.Drivers.Path)
			if err != nil {
				klog.Error("can not expand current node volume, ", err, ", ", node.Drivers.Path)
				return err
			}

			if _, pattern, ok := validateFirstAddingVolume(volumes[len(volumes)-1][0], base.MINIO_NODE_VOLUME); !ok {
				return errors.New("invalid volume to be added")
			} else {
				volumeEnd = currentVolumesEnd + 1
				replceVolume = fmt.Sprintf(pattern, fmt.Sprintf("{%d...%d}", currentVolumesStart, volumeEnd))
			}
		}

		replaceCurrentVolum, err := minio.ReplaceVolumeIndex(node.Drivers.Path, currentVolumesStart, volumeEnd)
		if err != nil {
			return err
		}

		if replaceCurrentVolum != replceVolume {
			klog.Error("different volume to be added is not supported")
			return errors.New("invalid volume")
		}

		node.Drivers.Num = (volumeEnd - currentVolumesStart + 1)
		node.Drivers.Path = replceVolume
	}

	base.MINIO_NODE_VOLUME = node.Drivers.Path
	if err = c.CreateOperatorServerEnv(key); err != nil {
		return err
	}

	err = nodeData.MarshalFrom(&node)
	if err != nil {
		klog.Error("encode node data error, ", err)
		return err
	}

	klog.Info("updating cluster config")
	err = c.BaseCommand.Client.Update(ctx, key, data.Modified, nodeData)
	if err != nil {
		klog.Error("update cluster config error, ", err, ", ", string(nodeData))
	}

	klog.Info("success to add driver to cluster, ", base.MINIO_NODE_VOLUME)
	return nil
}

func validateFirstAddingVolume(path, firstVolume string) (int, string, bool) {
	for i := 0; i < len(path); i++ {
		if path[i] != firstVolume[i] {
			if path[i] == '9' {
				if len(firstVolume) < (i + 2) {
					// the index cannot be 10
					return -1, "", false
				}

				s := firstVolume[i : i+2]
				index, err := strconv.Atoi(s)
				if err != nil {
					return -1, "", false
				}

				if index == 10 {
					return 9, "", true
				}

				return -1, "", false
			} // prev volume index 9

			if _, err := strconv.Atoi(path[i : i+1]); err != nil {
				return -1, "", false
			}

			if _, err := strconv.Atoi(firstVolume[i : i+1]); err != nil {
				return -1, "", false
			}

			if (firstVolume[i] - path[i]) == 1 {
				index, _ := strconv.Atoi(path[i : i+1])
				return index, path[:i] + "%s" + path[i+1:], true
			}

			return -1, "", false
		} // first different byte
	}

	return -1, "", false
}

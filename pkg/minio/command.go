package minio

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"bytetrade.io/web3os/minio-operator/pkg/utils"
	hostsfile "github.com/kevinburke/hostsfile/lib"
	"k8s.io/klog/v2"
)

func (config *ClusterConfig) CreateDefaultEnv() error {
	MinIOEnv["MINIO_ROOT_PASSWORD"] = config.RootPassword
	MinIOEnv["MINIO_ROOT_USER"] = config.RootUser

	switch config.Mode {
	case SNSD, SNMD:
		MinIOEnv["MINIO_VOLUMES"] = config.Nodes[0].Drivers.Path
		MinIOEnv["MINIO_SERVER_URL"] = fmt.Sprintf("http://%s:%s", config.Nodes[0].IP, config.Nodes[0].Port)
	case MNMD:
		volumes := []string{}

		for _, n := range config.Nodes {
			pool := fmt.Sprintf("http://%s:%s%s", n.Hostname, n.Port, n.Drivers.Path)
			volumes = append(volumes, pool)
		}

		MinIOEnv["MINIO_VOLUMES"] = strings.Join(volumes, " ")
		MinIOEnv["MINIO_SERVER_URL"] = fmt.Sprintf("http://%s:%s", config.Nodes[0].IP, config.Nodes[0].Port)
	}

	err := utils.WritePropertiesFile(MinIOEnv, DefaultConfigFile)
	if err != nil {
		klog.Error("write minio env file error, ", err, ", ", DefaultConfigFile)
	}

	return err
}

func (config *ClusterConfig) UpdateHosts() error {
	f, err := os.Open("/etc/hosts")
	if err != nil {
		klog.Error("open /etc/hosts error, ", err)
		return err
	}

	hosts, err := hostsfile.Decode(f)
	if err != nil {
		klog.Error("decode /etc/hosts error, ", err)
		return err
	}

	for _, n := range config.Nodes {
		ip, err := net.ResolveIPAddr("ip", n.IP)
		if err != nil {
			klog.Error("resolve node ip error, ", err, ", ", n.IP)
			return err
		}

		hosts.Set(*ip, n.Hostname)
	}

	tmpf, err := os.CreateTemp("/tmp", "hostsfile-temp")
	if err != nil {
		klog.Error("create hosts tmp file error, ", err)
		return err
	}

	err = hostsfile.Encode(tmpf, hosts)
	if err != nil {
		klog.Error("encode hosts tmp file error, ", err)
		return err
	}

	err = os.Chmod(tmpf.Name(), 0644)
	if err != nil {
		klog.Error("change hosts tmp file mode error, ", err)
		return err
	}

	err = os.Rename(tmpf.Name(), "/etc/hosts")
	if err != nil {
		klog.Error("mv hosts tmp file error, ", err)
		return err
	}

	return nil
}

func ClearMinIOSysfile(volumes string) error {
	if !strings.HasPrefix(volumes, "/") {
		klog.Error("invalid volumes to clear, ", volumes)
		return errors.New("invalid volumes")
	}

	expanded, err := VolumeExpend(volumes)
	if err != nil {
		klog.Error("expand volumes error, ", err, ", ", volumes)
		return err
	}

	for _, p := range expanded {
		if len(p) > 1 {
			continue
		}

		err = utils.RunOSCommand("rm", "-rf", strings.Join([]string{p[0], ".minio.sys"}, "/"))
		if err != nil {
			klog.Error("clear minio sys files error, ", err, ", ", strings.Join([]string{p[0], ".minio.sys"}, "/"))
			return err
		}
	}

	return nil
}

func ReloadMinIOOperatorService() error {
	return utils.ReloadSystemService("minio-operator")
}

func ReloadMinIOService() error {
	return utils.ReloadSystemService("minio")
}

package minio

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"bytetrade.io/web3os/minio-operator/pkg/storage/model"
	"github.com/magiconair/properties"
	"github.com/minio/pkg/v2/ellipses"
	"k8s.io/klog/v2"
)

const (
	DefaultConfigFile      = "/etc/default/minio"
	DefaultOperatorEnvFile = "/etc/default/minio-operator"
	OperatorServiceFile    = "/etc/systemd/system/minio-operator.service"
	MinIOServiceFile       = "/etc/systemd/system/minio.service"
	EnvVolumes             = "MINIO_VOLUMES"
	DefaultRootUser        = "minioadmin"
)

type ClusterMode string

const (
	SNSD ClusterMode = "SNSD"
	SNMD ClusterMode = "SNMD"
	MNMD ClusterMode = "MNMD"
)

type ClusterConfig struct {
	Mode         ClusterMode
	RootUser     string
	RootPassword string
	Nodes        []*model.Node
}

func GetCurrentMode() (ClusterMode, error) {
	p, err := properties.LoadFile(DefaultConfigFile, properties.UTF8)
	if err != nil {
		klog.Error("load default env config file error, ", err)
		return "", err
	}

	volume := strings.TrimSpace(p.GetString(EnvVolumes, ""))
	if volume == "" {
		klog.Error("volume define not found")
		return "", errors.New("volume not found")
	}

	volumes := strings.Split(volume, " ")

	// we just check the first endpoint configuration in MINIO_VOLUMES
	// cause if the first endpoint is invalid, the cluster will get to wrong state
	expanded, err := VolumeExpend(volumes[0])
	if err != nil {
		return "", err
	}

	klog.Info("expanded volume config: ", expanded)

	// if multi endpoints in configuration, the first endpoint must be MNMD
	mode, err := GetClusterMode(expanded)
	if err != nil {
		return "", err
	}

	switch mode {
	case SNMD, SNSD:
		if len(volumes) > 1 {
			return "", errors.New(EnvVolumes + " config invalid, " + volume)
		}
	}

	return mode, nil
}

func VolumeExpend(endpoint string) ([][]string, error) {
	if !ellipses.HasEllipses(endpoint) {
		return [][]string{{endpoint}}, nil
	}

	patterns, err := ellipses.FindEllipsesPatterns(endpoint)
	if err != nil {
		klog.Error("parse volume pattern error, ", err, ", ", endpoint)
		return nil, err
	}

	return patterns.Expand(), err
}

func GetClusterMode(expandedPattern [][]string) (ClusterMode, error) {
	if len(expandedPattern) == 1 && len(expandedPattern[0]) == 1 {
		return SNSD, nil
	}

	// multi driver with local volume
	if len(expandedPattern[0]) == 1 && strings.HasPrefix(expandedPattern[0][0], "/") {
		return SNMD, nil
	}

	getNode := func(endpoint string) (string, error) {
		nodeUrl, err := url.Parse(endpoint)
		if err != nil {
			klog.Error("parse endpoint url error, ", err, ", ", endpoint)
			return "", err
		}

		return nodeUrl.Host, nil
	}

	node, err := getNode(expandedPattern[0][0])
	if err != nil {
		return "", err
	}

	for _, ep := range expandedPattern {
		n, err := getNode(ep[0])
		if err != nil {
			return "", err
		}

		if n != node {
			return MNMD, nil
		}
	}

	return SNMD, nil
}

func NodeName(index int) string {
	return fmt.Sprintf("minio-%d", index)
}

func NodesName(first, end int) string {
	return fmt.Sprintf("minio-{%d...%d}", first, end)
}

func FindVolumeIndex(vol string) (start, end int, err error) {
	r, err := regexp.Compile(`[^{]*{([1-9]+)\.\.\.([1-9]*)}[^{]*`)
	if err != nil {
		klog.Error("regexp error, ", err)
		return
	}

	m := r.FindStringSubmatch(vol)

	if len(m) < 3 {
		klog.Error("match error, ", m)
		err = errors.New("invalid volume format")
		return
	}

	start, err = strconv.Atoi(m[1])
	if err != nil {
		klog.Error("volume start error, ", err, ", ", m[1])
		return
	}

	end, err = strconv.Atoi(m[2])
	if err != nil {
		klog.Error("volume end error, ", err, ", ", m[2])
		return
	}

	return
}

func ReplaceVolumeIndex(src string, start, end int) (string, error) {
	r, err := regexp.Compile(`([^{]*{)([1-9]+)(\.\.\.)([1-9]*)(}[^{]*)`)
	if err != nil {
		klog.Error("regexp error, ", err)
		return "", err
	}

	pattern := r.ReplaceAllString(src, "$1%d$3%d$5")

	return fmt.Sprintf(pattern, start, end), nil
}

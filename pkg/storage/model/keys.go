package model

import "strings"

const KEY_PREFIX = "terminus/minio/nodes"

const CLUSTER_KEY = "terminus/minio/cluster"

func GetNodeKey(nodeName string) string {
	return strings.Join([]string{KEY_PREFIX, nodeName}, "/")
}

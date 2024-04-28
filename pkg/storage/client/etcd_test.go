package client

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"bytetrade.io/web3os/minio-operator/pkg/storage/model"
	"k8s.io/klog/v2"
)

func getConfig() ETCDConfig {
	return ETCDConfig{
		Endpoints: []string{"192.168.50.50:2379"},
		TLSConfig: Config{
			CAFile:   "/Users/liuyu/workspace/ca.pem",
			CertFile: "/Users/liuyu/workspace/node-pengpeng-dev.pem",
			KeyFile:  "/Users/liuyu/workspace/node-pengpeng-dev-key.pem",
		},
	}
}

func TestWatch(t *testing.T) {
	config := getConfig()

	client, err := New(config)
	if err != nil {
		klog.Error(err)
		t.Fail()
		return
	}

	klog.Info("start to watch")
	defer client.Close()
	for watchResp := range client.Watch(context.Background(), "minio/watch") {
		for _, event := range watchResp.Events {
			fmt.Printf("Event received! %s executed on %q with value %q\n", event.Type, event.Kv.Key, event.Kv.Value)
		}
	}
}

func TestCreate(t *testing.T) {
	config := getConfig()
	client, err := New(config)
	if err != nil {
		klog.Error(err)
		t.Fail()
		return
	}

	defer client.Close()
	node := model.Node{
		Hostname: "minio-1",
		Drivers: model.Drivers{
			Path: "/vol",
			Num:  1,
		},
	}

	value, err := json.Marshal(node)
	if err != nil {
		klog.Error(err)
		t.Fail()
		return
	}

	err = client.Create(context.Background(), "minio/cluster", []byte(value))
	if err != nil {
		klog.Error(err)
		t.Fail()
		return
	}

	t.Log("create success")
}

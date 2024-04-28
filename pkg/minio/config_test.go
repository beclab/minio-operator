package minio

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"k8s.io/klog/v2"
)

func TestVolumeExpand(t *testing.T) {
	// v := "http://minio-{1...2}/terminus/data/minio/vol1"
	// v := "http://minio/terminus/data/minio/vol{1...2}"
	// v := "http://minio-{1...2}/terminus/data/minio/vol{1...2}"
	// v := "http://minio/terminus/data/minio/vol1"
	v := "/data/minio/vol{1...4}"

	p, e := VolumeExpend(v)
	if e != nil {
		t.Fail()
		return
	}

	t.Log("result: ", p)
}

func TestLstat(t *testing.T) {
	info, _ := os.Stat("/tmp/bin")
	fmt.Print(info.Mode()&os.ModeSymlink != 0)
}

func TestCurrentMode(t *testing.T) {
	mode, err := GetCurrentMode()
	if err != nil {
		klog.Error(err)
		t.Fail()
		return
	}

	t.Log(mode)
}

func TestVolume(t *testing.T) {
	v := "/data/vol/data{1...2}"

	r, err := regexp.Compile(`([^{]*{)([1-9]+)(\.\.\.)([1-9]*)(}[^{]*)`)
	if err != nil {
		klog.Error(err)
		t.Fail()
		return
	}

	m := r.FindStringSubmatch(v)
	t.Log(m)

	t.Log(r.ReplaceAllString(v, "$1%d$3%d$5"))
}

func TestClusterMode(t *testing.T) {
	v := "/terminus/data/minio/vol1"
	p, e := VolumeExpend(v)
	if e != nil {
		t.Fail()
		return
	}

	m, e := GetClusterMode(p)
	if e != nil {
		t.Fail()
		return
	}

	t.Log(m)
}

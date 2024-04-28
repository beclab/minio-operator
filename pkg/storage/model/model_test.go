package model

import (
	"testing"

	"k8s.io/klog/v2"
)

func TestSerialize(t *testing.T) {
	d := Drivers{Path: "/v1"}

	data := DataValue("sfskjhfskdjfhskdjhfskdjfhksdjfhksdfhksdjfhksdjhfkd")
	err := data.MarshalFrom(&d)
	if err != nil {
		klog.Error(err)
		t.Fail()
		return
	}

	klog.Info(string(data))

	var d1 Drivers

	err = data.UnmarshalTo(&d1)
	if err != nil {
		klog.Error(err)
		t.Fail()
		return
	}

	t.Log(d1.Path)
}

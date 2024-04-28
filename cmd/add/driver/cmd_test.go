package driver

import "testing"

func TestValidate(t *testing.T) {
	s, p, ok := validateFirstAddingVolume("/terminus/data1/minio/vol1", "/terminus/data2/minio/vol1")

	t.Log("result: ", s, p, ok)
}

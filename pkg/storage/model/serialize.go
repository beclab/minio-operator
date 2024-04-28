package model

import "encoding/json"

type DataValue []byte

func (d *DataValue) UnmarshalTo(data any) error {
	return json.Unmarshal(*d, data)
}

func (d *DataValue) MarshalFrom(data any) error {
	v, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(*d) < len(v) {
		*d = append(*d, v[len(*d):]...)
	} else if len(*d) > len(v) {
		*d = (*d)[:len(v)]
	}
	copy(*d, v)
	return nil
}

package modbus

import (
	"fmt"
	"reflect"
)

func Fuzz(data []byte) int {
	dec := &rtuPackager{}
	pdu, err := dec.Decode(data)
	if err != nil {
		if pdu != nil {
			panic("pdu != nil on error!")
		}
		return 0
	}
	adu, err := dec.Encode(pdu)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(adu, data) {
		panic(fmt.Errorf("two data should equal,%v %v", adu, data))
	}
	return 1
}

package proto

import (
	"errors"
	"reflect"
)

func ProtoSlice(src interface{}, dst interface{}) error {
	vsrc := reflect.ValueOf(src)
	if vsrc.Kind() != reflect.Slice {
		return errors.New("need slice")
	}

	vdst := reflect.ValueOf(dst)
	if vdst.Kind() != reflect.Ptr {
		return errors.New("dst need to be address")
	}

	tsrc := reflect.TypeOf(src)
	tdst := reflect.TypeOf(dst)
	typsrcelem := tsrc.Elem()
	typdstslice := tdst.Elem()
	typdstelem := typdstslice.Elem() // dst is a slice pointer

	m, found := typsrcelem.MethodByName("ToProto")
	if !found {
		return errors.New("no ToProto method")
	}

	if m.Type.NumOut() != 1 {
		return errors.New("incorrect method out num")
	}

	typsrcelem = m.Type.Out(0)
	if typsrcelem != typdstelem {
		return errors.New("out type not match")
	}

	rdst := reflect.MakeSlice(typdstslice, vsrc.Len(), vsrc.Len())
	for i := 0; i < vsrc.Len(); i++ {
		itsrc := vsrc.Index(i)

		callp := []reflect.Value{itsrc}
		result := m.Func.Call(callp)
		if result == nil || len(result) == 0 {
			return errors.New("ToProto failed")
		}

		rdst.Index(i).Set(result[0])
	}

	vdst.Elem().Set(rdst)
	return nil
}

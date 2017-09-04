package dqrack

import "reflect"

type walkerFunc = func(reflect.StructField, reflect.Value) error

func structWalker(s reflect.Value, fn walkerFunc) error {
	if s.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	ty := s.Type()
	fc := s.NumField()
	for i := 0; i < fc; i++ {
		sf := ty.Field(i)
		v := s.Field(i)

		// this is true if unexported
		if sf.PkgPath != "" {
			continue
		}

		err := fn(sf, v)
		if err != nil {
			return err
		}
	}

	return nil
}

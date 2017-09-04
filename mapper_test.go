package dqrack

import "testing"
import "reflect"

func TestStructWalker(t *testing.T) {
	type testStruct struct {
		Field1 bool
		Field2 bool
		Field3 bool
		Field4 bool
	}

	fieldsVisited := []string{}
	fieldsRequired := []string{"Field1", "Field2", "Field3", "Field4"}

	v := reflect.ValueOf(testStruct{})
	err := structWalker(v, func(sf reflect.StructField, v reflect.Value) error {
		fieldsVisited = append(fieldsVisited, sf.Name)
		return nil
	})
	if err != nil {
		t.Error(err)
		return
	}

	for _, v := range fieldsRequired {
		if !sliceContains(fieldsVisited, v) {
			t.Error("didn't visit", v)
		}
	}
}

func sliceContains(s []string, v string) bool {
	for _, val := range s {
		if val == v {
			return true
		}
	}

	return false
}

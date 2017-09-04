package dqrack

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/dgraph-io/dgraph/protos"
)

func (dq *Dqrack) Scan(identity string, v interface{}) error {
	edges := []string{}
	edgeVals := map[string]reflect.Value{}

	rv := reflect.ValueOf(v).Elem()

	structWalker(rv, func(sf reflect.StructField, iv reflect.Value) error {
		dq.d("walking", sf)
		key := sf.Name

		dk := structKeyToName(key)

		skey := getTag(sf, dk)
		to := parseDqTag(skey)

		// handle ex. dq:"-"
		if to.Ignore {
			return nil
		}

		// handle ex. dq:",index"
		if to.Name == "" {
			to.Name = dk
		}

		edgeVals[to.Name] = iv
		edges = append(edges, to.Name)
		return nil
	})

	dq.d("getting edges", edges)

	o, err := dq.getValues(fmt.Sprintf(`eq(_identity, "%s")`, identity), edges)
	if err != nil {
		return err
	}

	if len(o.Children) == 0 {
		return ErrEmpty
	}

	for _, prop := range o.Children[0].Properties {
		v, ok := edgeVals[prop.Prop]
		if !ok {
			dq.d("skipped unknown", prop)
			continue
		}

		err := dq.getEdge(prop.Value, &v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dq *Dqrack) getEdge(e *protos.Value, v *reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.String:
		v.SetString(e.GetStrVal())
	case reflect.Bool:
		v.SetBool(e.GetBoolVal())
	case reflect.Int:
		v.SetInt(e.GetIntVal())
	default:
		val := e.GetBytesVal()
		i := v.Interface()

		err := json.Unmarshal(val, &i)
		if err != nil {
			return err
		}

		v.Set(reflect.ValueOf(i))
	}

	return nil
}

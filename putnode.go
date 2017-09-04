package dqrack

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	dgraph "github.com/dgraph-io/dgraph/client"
	"github.com/imdario/mergo"
)

// PutNode creates a node with edges from the struct.
func (dq *Dqrack) PutNode(v Qrackable) (n dgraph.Node, err error) {
	vName := v.GetName()
	n, err = dq.Dgraph.NodeBlank(vName)
	if err != nil {
		return
	}

	var d interface{}
	d = v

	rd := reflect.ValueOf(v)

	gd := rd.MethodByName("GetData")
	if gd.IsValid() {
		d = gd.Call([]reflect.Value{})[0].Interface()
	}

	rt := reflect.TypeOf(d)
	if rt.Kind() != reflect.Struct {
		return n, ErrNotStruct
	}

	var typeName string
	gt := rd.MethodByName("GetType")
	if gt.IsValid() {
		typeName = gt.Call([]reflect.Value{})[0].String()
	} else {
		typeName = reflect.TypeOf(v).Name()
	}

	if typeName == "" {
		return n, ErrEmptyTypeName
	}

	rv := reflect.ValueOf(d)
	// fm := dq.mapper.FieldMap(rv)

	dq.d("node", vName, "of", typeName)
	req := &dgraph.Req{}

	var e dgraph.Edge

	// basic edges
	// type makes it searchable by the struct name
	e = n.Edge("_type")
	e.SetValueString(strings.ToLower(typeName))
	err = req.Set(e)
	if err != nil {
		return
	}

	// identity is it's own name, or specific identity.
	e = n.Edge("_identity")
	e.SetValueString(vName)
	err = req.Set(e)
	if err != nil {
		return
	}

	err = dq.walk(n, req, rv, "")
	if err != nil {
		return n, fmt.Errorf("walk: %v", err)
	}

	_, err = dq.Run(req)
	if err != nil {
		return n, fmt.Errorf("run: %v", err)
	}

	dq.lru.Add(vName, n)

	return
}

func (dq *Dqrack) walk(n Node, req *dgraph.Req, rv reflect.Value, prefix string) error {
	return structWalker(rv, func(sf reflect.StructField, v reflect.Value) error {
		key := sf.Name

		dk := structKeyToName(key)

		skey := getTag(sf, dk)
		to := parseDqTag(skey)

		// handle ex. dq:"-"
		if to.Ignore {
			return nil
		}

		dq.d(to)

		// handle ex. dq:",index"
		if to.Name == "" {
			to.Name = dk
		}

		to.Name = prefix + to.Name

		dq.d("in", key, "dq:", skey, "to:", to.Name)

		if to.Inline && v.Type().Kind() == reflect.Struct {
			// recurse here
			dq.d("going into", to.Name)

			np := ""

			if to.Prefix {
				np = to.Name + "_"
			}

			return dq.walk(n, req, v, np)
		}

		e := n.Edge(to.Name)
		err := dq.setEdge(&e, v)
		if err != nil {
			return fmt.Errorf("setEdge: %v", err)
		}
		err = req.Set(e)
		if err != nil {
			return fmt.Errorf("set: %v", err)
		}

		return nil
	})
}

// setEdge figures out the best edge type for some struct value
// TODO: forcibly define default with struct tags.
func (dq *Dqrack) setEdge(e *dgraph.Edge, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.String:
		dq.d("putting string", v.String())
		val := strings.Replace(v.String(), "\"", "\\\"", -1)
		if val == "" {
			return ErrEmpty
		}
		return e.SetValueString(val)
	case reflect.Bool:
		dq.d("putting bool", v.Bool())
		return e.SetValueBool(v.Bool())
	case reflect.Int:
		dq.d("putting int", v.Int())
		return e.SetValueInt(v.Int())
	default:
		b, err := json.Marshal(v.Interface())
		if err != nil {
			return err
		}

		dq.d("defaulting", string(b))
		return e.SetValueBytes(b)
	}
}

func getTag(sf reflect.StructField, key string) (s string) {
	s = sf.Tag.Get("dq")
	if s != "" {
		return
	}

	s = sf.Tag.Get("json")
	if s != "" {
		return
	}

	return key
}

func parseDqTag(tag string) tagOpts {
	p := strings.Split(tag, ",")

	om := map[string]interface{}{}

	for _, k := range p[1:] {
		if k == "omitempty" {
			k = "OmitEmpty"
		}

		om[k] = true
	}

	t := tagOptsFromMap(om)
	t.Name = p[0]

	if t.Name == "-" {
		t.Ignore = true
	}

	return t
}

func tagOptsFromMap(om map[string]interface{}) (to tagOpts) {
	mergo.Map(&to, om)
	return
}

func structKeyToName(key string) string {
	s := strings.ToLower(key)
	s = strings.Replace(s, ".", "_", -1)
	return s
}

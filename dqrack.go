// Package dqrack is a helping wrapper for Dgraph.
// It provides a load of useful "marshalling" functions
// to allow lazy work.
//
// Pronounced "d-crack"
package dqrack // import "github.com/kayteh/dqrack"

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	dgraph "github.com/dgraph-io/dgraph/client"
	"github.com/imdario/mergo"
	"github.com/jmoiron/sqlx/reflectx"
)

// Dqrack is a set of cheap tricks, like any other cheap trick library.
type Dqrack struct {
	Dgraph *dgraph.Dgraph
	Mapper *reflectx.Mapper
}

type tagOpts struct {
	// Name of key
	Name string

	// Don't add the node if this is considered empty
	OmitEmpty bool

	// Create facets from the sub-struct
	Facets bool

	// Flatten sub-structs into the outer
	Inline bool

	// for `dq:"key,inline,prefix"`, we flatten the structure like key_subkey
	Prefix bool

	// Skip marshalling this, only on `dq:"-"`
	Ignore bool
}

// Qrackable types let us get information out of structs in a non-arbitrary fashion.
type Qrackable interface {
	GetName() string
	GetData() interface{}
}

func New(dg *dgraph.Dgraph) *Dqrack {
	return &Dqrack{
		Dgraph: dg,
		Mapper: reflectx.NewMapper("dq"),
	}
}

// GetNode creates a node with edges from the struct.
func (dq *Dqrack) GetNode(v Qrackable) (n dgraph.Node, err error) {
	vName := v.GetName()
	n, err = dq.Dgraph.NodeBlank(vName)
	if err != nil {
		return
	}

	d := v.GetData()
	// d := v

	rt := reflect.TypeOf(d)
	qt := reflect.TypeOf(v)
	rv := reflect.ValueOf(d)
	fm := dq.Mapper.FieldMap(rv)

	log.Println("node", vName, "of", qt.Name())
	req := &dgraph.Req{}

	var e dgraph.Edge

	// basic edges
	// type makes it searchable by the struct name
	e = n.Edge("_type")
	e.SetValueString(strings.ToLower(qt.Name()))
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

	for key, rfv := range fm {
		if rfv.Type().Kind() == reflect.Struct {
			continue
		}

		sf, _ := rt.FieldByName(key)
		skey := getTag(sf, key)
		to := parseDqTag(skey)

		if to.Ignore {
			continue
		}

		log.Println("in", key, "dq:", skey)

		e = n.Edge(to.Name)
		err = setEdge(&e, rfv)
		if err != nil {
			return n, fmt.Errorf("setEdge: %v", err)
		}
		err = req.Set(e)
		if err != nil {
			return n, fmt.Errorf("set: %v", err)
		}
	}

	_, err = dq.Dgraph.Run(context.Background(), req)
	if err != nil {
		return n, fmt.Errorf("run: %v", err)
	}

	return
}

// setEdge figures out the best edge type for some struct value
// TODO: forcibly define default with struct tags.
func setEdge(e *dgraph.Edge, v reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.String:
		log.Println("putting string", v.String())
		val := strings.Replace(v.String(), "\"", "\\\"", -1)
		return e.SetValueString(val)
	case reflect.Bool:
		log.Println("putting bool", v.Bool())
		return e.SetValueBool(v.Bool())
	case reflect.Int:
		log.Println("putting int", v.Int())
		return e.SetValueInt(v.Int())
	default:
		b, err := json.Marshal(v.Interface())
		if err != nil {
			return err
		}

		log.Println("defaulting", string(b))
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

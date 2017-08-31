package dqrack

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	dgraph "github.com/dgraph-io/dgraph/client"
)

func TestGetNode(t *testing.T) {
	req := &dgraph.Req{}
	req.SetQuery(`
		mutation {
			schema {
				_identity: string @index(exact) .
				_type: string @index(exact) .
			}
		}
	`)
	_, err := dq.Dgraph.Run(context.Background(), req)
	if err != nil {
		t.Error(err)
		return
	}

	n := testTnode
	n.identity = n.GetName()
	_, err = dq.GetNode(n)
	if err != nil {
		t.Error(err)
		return
	}

	// time.Sleep(3 * time.Second)

	o, err := getValues(`eq(_identity, "`+n.identity+`")`, []string{"f1"})
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(o.String())
	if len(o.Children[0].Properties) == 0 {
		t.Error("props was zero. bad.", o)
	}
}

func TestSetEdge(t *testing.T) {
	n, err := dq.Dgraph.NodeBlank("test_set_edge")
	if err != nil {
		t.Error(err)
		return
	}

	edgeName := fmt.Sprintf("testedge_%d", rand.Int31())

	req := &dgraph.Req{}
	req.SetQuery(`
		mutation {
			schema {
				` + edgeName + `: string @index(exact) .
			}
		}
	`)

	e := n.Edge(edgeName)
	err = setEdge(&e, reflect.ValueOf(0))
	if err != nil {
		t.Error(err)
		return
	}
	err = setEdge(&e, reflect.ValueOf(true))
	if err != nil {
		t.Error(err)
		return
	}
	err = setEdge(&e, reflect.ValueOf(map[string]string{"waifu": "you"}))
	if err != nil {
		t.Error(err)
		return
	}
	err = setEdge(&e, reflect.ValueOf("testscalar"))
	if err != nil {
		t.Error(err)
		return
	}
	req.Set(e)
	fmt.Println("done setting")

	_, err = dq.Dgraph.Run(context.Background(), req)
	if err != nil {
		t.Error(err)
		return
	}

	o, err := getValues(`eq(`+edgeName+`, "testscalar")`, []string{edgeName})
	if err != nil {
		t.Error(err)
		return
	}

	if len(o.Children[0].Properties) == 0 {
		t.Error("props was zero. bad.", o)
	}
}

func TestSetEdgeErr_Quotes(t *testing.T) {
	n, err := dq.Dgraph.NodeBlank("test_set_edge_err_quotes")
	if err != nil {
		t.Error(err)
		return
	}

	edgeName := fmt.Sprintf("testedge_%d", rand.Int31())
	e := n.Edge(edgeName)
	err = setEdge(&e, reflect.ValueOf(`"this has quotes!"`))
	if err != nil {
		t.Error(err)
		return
	}
	req := &dgraph.Req{}
	req.Set(e)

	_, err = dq.Dgraph.Run(context.Background(), req)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestTagParser(t *testing.T) {
	// parse tags like dq:"someKey,etc" into component parts
	to := parseDqTag("key,omitempty,inline,prefix")
	if to.Name != "key" {
		t.Error("tag wasn't parsed correctly:", to)
		return
	}

	fmt.Println(to)

	to = parseDqTag("-,omitempty,inline,prefix")
	if !to.Ignore {
		t.Error("tag wasn't parsed correctly to ignore:", to)
		return
	}

	fmt.Println(to)
}

func TestKeyToName(t *testing.T) {
	e := structKeyToName("ICant.CSharp")
	if e != "icant_csharp" {
		t.Error("unexpected value, got", e)
	}
}

func TestStructFieldKey(t *testing.T) {
	type testStr struct {
		Test     int `dq:"tagtagtag" json:"blah2"`
		TestJSON int `json:"blah"`
		TestNone int
	}

	ts := reflect.TypeOf(testStr{})
	sf0 := ts.Field(0)
	sf1 := ts.Field(1)
	sf2 := ts.Field(2)

	if getTag(sf0, "sf0") != "tagtagtag" {
		t.Error("wrong val on sf0")
	}

	if getTag(sf1, "sf1") != "blah" {
		t.Error("wrong val on sf1")
	}

	if getTag(sf2, "sf2") != "sf2" {
		t.Error("wrong val on sf2")
	}
}

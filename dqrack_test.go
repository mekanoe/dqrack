package dqrack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/segmentio/ksuid"

	dgraph "github.com/dgraph-io/dgraph/client"
	"github.com/dgraph-io/dgraph/protos"
	"google.golang.org/grpc"
)

var dq *Dqrack

func TestMain(m *testing.M) {
	g, err := grpc.Dial(os.Getenv("DGRAPH_ADDR"), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	dg := dgraph.NewDgraphClient([]*grpc.ClientConn{g}, dgraph.DefaultOptions, "/tmp/.tmp-"+time.Now().String())

	dq = New(dg)

	jsonCast(&testTnode, map[string]interface{}{
		"Field1": "hello!",
		"Sub": map[string]string{
			"HiHi": "hi!",
		},
	})

	os.Exit(m.Run())
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

type Tnode struct {
	Field1 string `dq:"f1"`
	Sub    struct {
		HiHi string `dq:"f4"`
	}

	identity string
}

var testTnode Tnode

func (t Tnode) GetName() string {
	if t.identity == "" {
		t.identity = ksuid.New().String()
	}

	return t.identity
}

func (t Tnode) GetData() interface{} {
	return t
}

func (t Tnode) GetType() string {
	return "Tnode"
}

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

func getValues(search string, preds []string) (n *protos.Node, err error) {
	req := &dgraph.Req{}
	req.SetQuery(`
		{
			qq(func: ` + search + `) {
				` + strings.Join(preds, "\n") + `
				_predicate_
			}
		}
	`)

	r, err := dq.Dgraph.Run(context.Background(), req)
	if err != nil {
		return
	}

	return r.N[0], err
}

// Helper func for casting weird types into good types.
func jsonCast(dst, src interface{}) error {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(src)
	if err != nil {
		return err
	}

	return json.NewDecoder(buf).Decode(dst)
}

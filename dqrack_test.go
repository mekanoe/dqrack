package dqrack

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/segmentio/ksuid"

	dgraph "github.com/dgraph-io/dgraph/client"
	"github.com/dgraph-io/dgraph/protos"
	"google.golang.org/grpc"
)

var tDq *Dqrack

func TestMain(m *testing.M) {
	g, err := grpc.Dial(os.Getenv("DGRAPH_ADDR"), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	dg := dgraph.NewDgraphClient([]*grpc.ClientConn{g}, dgraph.DefaultOptions, "/tmp/.tmp-"+time.Now().String())

	tDq, _ = New(dg)
	tDq.Debug = true

	jsonCast(&testTnode, map[string]interface{}{
		"Field1": "hello!",
		"Sub": map[string]string{
			"HiHi": "hi!",
		},
	})

	os.Exit(m.Run())
}

type Tnode struct {
	Field1 string `dq:"f1"`
	Sub    struct {
		HiHi string `dq:"f4"`
	} `dq:",inline,prefix"`

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

func getValues(search string, preds []string) (n *protos.Node, err error) {
	return tDq.getValues(search, preds)
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

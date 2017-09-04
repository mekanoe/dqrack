package dqrack

import (
	"testing"
)

func TestScan(t *testing.T) {
	var cacheNode Tnode
	jsonCast(&cacheNode, map[string]interface{}{
		"Field1": "hello!",
		"Sub": map[string]string{
			"HiHi": "hi!",
		},
	})
	cacheNode.identity = cacheNode.GetName()

	_, err := tDq.PutNode(cacheNode)
	if err != nil {
		t.Error(err)
		return
	}

	var outNode Tnode
	err = tDq.Scan(cacheNode.identity, &outNode)
	if err != nil {
		t.Error(err)
		return
	}

	if outNode.Field1 != cacheNode.Field1 {
		t.Error("nodes differ,", outNode, cacheNode)
	}
}

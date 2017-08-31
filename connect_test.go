package dqrack

import (
	"testing"
)

func TestConnect(t *testing.T) {
	var connNode1 Tnode
	jsonCast(&connNode1, map[string]interface{}{
		"Field1": "node1",
		"Sub": map[string]string{
			"HiHi": "hi!",
		},
	})
	connNode1.identity = connNode1.GetName()

	var connNode2 Tnode
	jsonCast(&connNode2, map[string]interface{}{
		"Field1": "node2",
		"Sub": map[string]string{
			"HiHi": "hi!",
		},
	})
	connNode2.identity = connNode2.GetName()

	_, err := dq.GetNode(connNode1)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = dq.GetNode(connNode2)
	if err != nil {
		t.Error(err)
		return
	}

	err = dq.ConnectNodes(connNode1, "next-to", connNode2)
	if err != nil {
		t.Error(err)
		return
	}

	o, err := dq.Fetch(connNode1, []string{"next-to { f1 }"})
	if err != nil {
		t.Error(err)
		return
	}

	if len(o.Children) == 0 {
		t.Error("children at 0, bad bad bad =>", o)
		return
	}

	aa := o.Children[0].Children[0].Properties[0].Value.GetStrVal()

	if aa != connNode2.Field1 {
		t.Error("nodes didn't connect or something")
	}
}

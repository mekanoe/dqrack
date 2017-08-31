package dqrack

import (
	dgraph "github.com/dgraph-io/dgraph/client"
)

func (dq *Dqrack) ConnectNodes(v1 Qrackable, pred string, v2 Qrackable) error {
	n1, err := dq.Node(v1)
	if err != nil {
		return err
	}

	n2, err := dq.Node(v2)
	if err != nil {
		return err
	}

	e := n1.ConnectTo(pred, n2)
	req := &dgraph.Req{}
	req.Set(e)
	_, err = dq.Run(req)

	return err
}

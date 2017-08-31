package dqrack

import (
	"fmt"
	"strings"

	dgraph "github.com/dgraph-io/dgraph/client"
	"github.com/dgraph-io/dgraph/protos"
)

func (dq *Dqrack) cacheGet(identity string) (n Node, err error) {
	if dq.lru.Contains(identity) {
		v, _ := dq.lru.Get(identity)
		n = v.(Node)
		return
	}

	r, err := dq.getValues(fmt.Sprintf(`eq(_identity, "%s")`, identity), []string{
		"_uid_",
	})
	if err != nil {
		return
	}

	if len(r.Children) == 0 {
		return
	}

	uid := r.Children[0].Properties[0].Value.GetUidVal()

	n = dq.Dgraph.NodeUid(uid)

	dq.lru.Add(identity, n)

	return n, nil
}

// Node fetches a node based on a Qrackable struct.
// An empty node can be tested for with n.String() == ""
func (dq *Dqrack) Node(v Qrackable) (n Node, err error) {
	return dq.cacheGet(v.GetName())
}

// Node fetches a node based on a Qrackable struct.
// An empty node can be tested for with n.String() == ""
func (dq *Dqrack) Fetch(v Qrackable, preds []string) (n *protos.Node, err error) {
	return getValues(fmt.Sprintf(`eq(_identity, "%s")`, v.GetName()), preds)
}

func (dq *Dqrack) getValues(search string, preds []string) (n *protos.Node, err error) {
	req := &dgraph.Req{}
	req.SetQuery(fmt.Sprintf(`
		{
			qq(func: %s) {
				%s
			}
		}
	`, search, strings.Join(preds, "\n")))

	r, err := dq.Run(req)
	if err != nil {
		return
	}

	return r.N[0], err
}

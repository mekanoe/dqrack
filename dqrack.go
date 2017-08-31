// Package dqrack is a helping wrapper for Dgraph.
// It provides a load of useful "marshalling" functions
// to allow lazy work.
//
// Pronounced "d-crack"
package dqrack // import "github.com/kayteh/dqrack"

import (
	"context"
	"errors"

	dgraph "github.com/dgraph-io/dgraph/client"
	"github.com/dgraph-io/dgraph/protos"
	"github.com/hashicorp/golang-lru"
	"github.com/jmoiron/sqlx/reflectx"
)

var (
	ErrEmptyTypeName = errors.New("type name was empty. this kills the batman.")
	ErrNotStruct     = errors.New("data was not a struct")
	ErrEmpty         = errors.New("empty value")
	ErrNil           = errors.New("argument was nil, expected not nil")
)

// Convenience alias for a dgraph Node
type Node = dgraph.Node

// Dqrack is a set of cheap tricks, like any other cheap trick library.
type Dqrack struct {
	Dgraph *dgraph.Dgraph

	mapper *reflectx.Mapper
	lru    *lru.ARCCache
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

	// Index in schema
	Index bool
}

// Qrackable types let us get information out of structs in a non-arbitrary fashion.
type Qrackable interface {
	GetName() string
}

type QDataable interface {
	GetData() interface{}
}

type QTypeable interface {
	GetType() string
}

func New(dg *dgraph.Dgraph) (*Dqrack, error) {
	l, err := lru.NewARC(1024)
	if err != nil {
		return nil, err
	}

	if dg == nil {
		return nil, ErrNil
	}

	d := &Dqrack{
		Dgraph: dg,
		mapper: reflectx.NewMapper("dqwwww"),
		lru:    l,
	}

	req := &dgraph.Req{}
	req.SetQuery(`
		mutation {
			schema {
				_identity: string @index(exact) .
				_type: string @index(exact) .
			}
		}
	`)
	_, err = d.Run(req)

	return d, err
}

func (dq *Dqrack) Run(req *dgraph.Req) (*protos.Response, error) {
	return dq.Dgraph.Run(context.Background(), req)
}

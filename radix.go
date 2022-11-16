package router

import (
	"fmt"
	"io"
)

// strcmp compares two strings and returns number of equal characters.
func strcmp(a, b string) int {
	if len(a) > len(b) {
		a, b = b, a
	}

	eq := 0
	for eq < len(a) && a[eq] == b[eq] {
		eq++
	}

	return eq
}

type radixEdge struct {
	prefix string
	node   *Radix
}

func newEdge(prefix string, value any) *radixEdge {
	return &radixEdge{
		prefix: prefix,
		node: &Radix{
			value: value,
		},
	}
}

// split splits edge on two parts.
// for example edge has a prefix "computer".
// on appends "command" to tree l will be 3
// so we need to split edge on two parts:
// 1. e.prefix becomes "com"
// 2. e.node moved to the new child edge with prefix "puter"
func (e *radixEdge) split(l int) {
	n := &radixEdge{
		prefix: e.prefix[l:],
		node:   e.node,
	}
	e.prefix = e.prefix[:l]
	e.node = new(Radix)
	e.node.edges = append(e.node.edges, n)
}

// Radix is a radix tree implementation.
type Radix struct {
	edges []*radixEdge
	value any
}

// find finds the edge that has the longest prefix match with the key.
// It returns the edge index, edge, and the number of equal characters.
func (n *Radix) find(key string) (int, *radixEdge, int) {
	for i := 0; i < len(n.edges); i++ {
		e := n.edges[i]
		eq := strcmp(e.prefix, key)
		if eq != 0 {
			return i, e, eq
		}
	}

	return -1, nil, 0
}

// Insert inserts a new value into the tree.
// If value is nil it removes the given key.
// Returns the old value and true if value was overwritten.
func (n *Radix) Insert(key string, value any) (any, bool) {
	if value == nil {
		return n.Remove(key)
	}

	if key == "" {
		if n.value == nil {
			n.value = value
			return nil, false
		} else {
			n.value, value = value, n.value
			return value, true
		}
	}

	_, e, eq := n.find(key)

	switch eq {
	case 0:
		// Edge not found
		n.edges = append(
			n.edges,
			newEdge(key, value),
		)

	case len(e.prefix):
		// The prefix is shorter than the edge prefix
		return e.node.Insert(key[eq:], value)

	case len(key):
		// The key is shorter than the prefix
		e.split(eq)
		e.node.value = value

	default:
		// The prefix has a mismatch with the key
		e.split(eq)
		e.node.edges = append(
			e.node.edges,
			newEdge(key[eq:], value),
		)
	}

	return nil, false
}

// Lookup finds the value for the given key.
func (n *Radix) Lookup(key string) (any, bool) {
	for key != "" {
		_, e, l := n.find(key)
		if e == nil || l != len(e.prefix) {
			return nil, false
		}

		n = e.node
		key = key[l:]
	}

	return n.value, true
}

// LookupPath finds the value for the given path.
// if equal path is not found it returns the root value with the longest prefix match.
//
// for example tree has next paths:
// - `/` - 404 page not found (root handler)
// - `/api` - api docs
// - `/api/users` - users list
// - `/api/users/` - 404 user not found (root handler)
// - `/api/users/admin` - user details
//
// next request will be handled by 404 page not found:
// - `/`
// - `/not-found`
// - `/api/`
//
// next request will be handled by 404 user not found:
// - `/api/users/`
// - `/api/users/not-found`
func (n *Radix) LookupPath(path string) any {
	lastRoot := n

	for path != "" {
		_, e, l := n.find(path)
		if e == nil || l != len(e.prefix) {
			return lastRoot.value
		}

		if e.node.value != nil && (e.prefix[l-1] == '/') {
			lastRoot = e.node
		}

		n = e.node
		path = path[l:]
	}

	return n.value
}

// Remove removes the value for the given path.
// Returns value and true if value was found and removed.
func (n *Radix) Remove(key string) (any, bool) {
	if key == "" {
		// node found. unset value and return
		if n.value == nil {
			// root node without value
			return nil, false
		}

		value := n.value
		n.value = nil
		return value, true
	}

	i, e, l := n.find(key)
	if e == nil || l != len(e.prefix) {
		// node not found
		return nil, false
	}

	value, ok := e.node.Remove(key[l:])

	if e.node.value == nil {
		switch len(e.node.edges) {
		case 0:
			// remove empty node
			last := len(n.edges) - 1
			n.edges[i] = n.edges[last]
			n.edges = n.edges[:last]
		case 1:
			// merge nodes
			last := e.node.edges[0]
			e.prefix += last.prefix
			e.node = last.node
		}
	}

	return value, ok
}

func (n *Radix) dump(out io.Writer, pad string) {
	last := len(n.edges) - 1

	for i := 0; i <= last; i++ {
		e := n.edges[i]

		if i != last {
			fmt.Fprintf(out, "%s├─── %s -> %v\n", pad, e.prefix, e.node.value)
			e.node.dump(out, (pad + "│    "))
		} else {
			fmt.Fprintf(out, "%s└─── %s -> %v\n", pad, e.prefix, e.node.value)
			e.node.dump(out, (pad + "     "))
		}
	}
}

// Dump prints the tree to out.
func (n *Radix) Dump(out io.Writer) {
	n.dump(out, "")
}

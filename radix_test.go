package router

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Edge returns edge with given prefix.
func (n *Radix) edge(prefix string) *radixEdge {
	for _, e := range n.edges {
		if e.prefix == prefix {
			return e
		}
	}

	return nil
}

func TestRadix_strcmp(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		a, b string
		want int
	}{
		{"water", "apple", 0},
		{"water", "water", 5},
		{"water", "watermelon", 5},
		{"watermelon", "water", 5},
		{"water", "wine", 1},
	}

	for _, test := range tests {
		if got := strcmp(test.a, test.b); !assert.Equal(test.want, got) {
			return
		}
	}
}

func TestRadix_split(t *testing.T) {
	assert := assert.New(t)

	edge := newEdge("computer", 1)
	edge.split(3)

	assert.Equal("com", edge.prefix)
	assert.Nil(edge.node.value)
	assert.Len(edge.node.edges, 1)

	edge = edge.node.edges[0]
	assert.Equal("puter", edge.prefix)
	assert.Equal(1, edge.node.value)
	assert.Len(edge.node.edges, 0)
}

func TestRadix_Insert(t *testing.T) {
	t.Run("base", func(t *testing.T) {
		assert := assert.New(t)

		// r
		// ├── om (8)
		// │   ├── ulus (3)
		// │   └── an
		// │       ├── e (1)
		// │       └── us (2)
		// └── ub
		//     ├── e
		//     │   ├── ns (4)
		//     │   └── r (5)
		//     └── ic
		//         ├── on (6)
		//         └── undus (7)
		currentNode := new(Radix)
		currentNode.Insert("romane", 1)
		currentNode.Insert("romanus", 2)
		currentNode.Insert("romulus", 3)
		currentNode.Insert("rubens", 4)
		currentNode.Insert("ruber", 5)
		currentNode.Insert("rubicon", 6)
		currentNode.Insert("rubicundus", 7)
		currentNode.Insert("rom", 8)

		currentNode.Dump(os.Stdout)

		assert.Nil(currentNode.value)
		assert.Len(currentNode.edges, 1)

		// level 1
		r := currentNode.edge("r")
		if !assert.NotNil(r) {
			return
		}
		assert.Nil(r.node.value)
		assert.Len(r.node.edges, 2)

		// level 2
		rom := r.node.edge("om")
		if !assert.NotNil(rom) {
			return
		}
		assert.Equal(8, rom.node.value)
		assert.Len(rom.node.edges, 2)

		rub := r.node.edge("ub")
		if !assert.NotNil(rub) {
			return
		}
		assert.Nil(rub.node.value)
		assert.Len(rub.node.edges, 2)

		// level 3
		if romulus := rom.node.edge("ulus"); assert.NotNil(romulus) {
			assert.Equal(3, romulus.node.value)
			assert.Len(romulus.node.edges, 0)
		}

		roman := rom.node.edge("an")
		if !assert.NotNil(roman) {
			return
		}
		assert.Nil(roman.node.value)
		assert.Len(roman.node.edges, 2)

		rube := rub.node.edge("e")
		if !assert.NotNil(rube) {
			return
		}
		assert.Nil(rube.node.value)
		assert.Len(rube.node.edges, 2)

		rubic := rub.node.edge("ic")
		if !assert.NotNil(rubic) {
			return
		}
		assert.Nil(rubic.node.value)
		assert.Len(rubic.node.edges, 2)

		// level 4
		if romane := roman.node.edge("e"); assert.NotNil(romane) {
			assert.Equal(1, romane.node.value)
			assert.Len(romane.node.edges, 0)
		}

		if romanus := roman.node.edge("us"); assert.NotNil(romanus) {
			assert.Equal(2, romanus.node.value)
			assert.Len(romanus.node.edges, 0)
		}

		if rubens := rube.node.edge("ns"); assert.NotNil(rubens) {
			assert.Equal(4, rubens.node.value)
			assert.Len(rubens.node.edges, 0)
		}

		if ruber := rube.node.edge("r"); assert.NotNil(ruber) {
			assert.Equal(5, ruber.node.value)
			assert.Len(ruber.node.edges, 0)
		}

		if rubicon := rubic.node.edge("on"); assert.NotNil(rubicon) {
			assert.Equal(6, rubicon.node.value)
			assert.Len(rubicon.node.edges, 0)
		}

		if rubicundus := rubic.node.edge("undus"); assert.NotNil(rubicundus) {
			assert.Equal(7, rubicundus.node.value)
			assert.Len(rubicundus.node.edges, 0)
		}
	})

	t.Run("insert duplicate", func(t *testing.T) {
		assert := assert.New(t)

		var updated bool
		currentNode := new(Radix)

		_, updated = currentNode.Insert("alert", 1)
		assert.False(updated)
		_, updated = currentNode.Insert("apple", 2)
		assert.False(updated)

		if v, updated := currentNode.Insert("alert", 3); assert.True(updated) {
			assert.Equal(1, v)
		}

		edge := currentNode.edge("a")
		assert.Len(edge.node.edges, 2)

		if e := edge.node.edge("pple"); !assert.NotNil(e) {
			assert.Equal(2, e.node.value)
		}
		if e := edge.node.edge("lert"); !assert.NotNil(e) {
			assert.Equal(3, e.node.value)
		}
	})
}

func TestRadix_LookupPath(t *testing.T) {
	assert := assert.New(t)

	currentNode := new(Radix)

	currentNode.Insert("/", 0)
	currentNode.Insert("/api", 1)
	currentNode.Insert("/api/users", 2)
	currentNode.Insert("/api/users/", 3)
	currentNode.Insert("/api/users/admin", 4)

	// lastRoot should be switched to the node with the longest prefix
	// with trailing slash, and with default value
	currentNode.Insert("/last-root/a", 11)
	currentNode.Insert("/last-root/b", 12)

	currentNode.Dump(os.Stdout)

	// defined paths
	if r := currentNode.LookupPath("/"); !assert.Equal(0, r) {
		return
	}
	if r := currentNode.LookupPath("/api"); !assert.Equal(1, r) {
		return
	}
	if r := currentNode.LookupPath("/api/users"); !assert.Equal(2, r) {
		return
	}
	if r := currentNode.LookupPath("/api/users/"); !assert.Equal(3, r) {
		return
	}
	if r := currentNode.LookupPath("/api/users/admin"); !assert.Equal(4, r) {
		return
	}

	// not found
	if r := currentNode.LookupPath("/not-found"); !assert.Equal(0, r) {
		return
	}
	if r := currentNode.LookupPath("/api/"); !assert.Equal(0, r) {
		return
	}
	if r := currentNode.LookupPath("/api/users/not-found"); !assert.Equal(3, r) {
		return
	}
	if r := currentNode.LookupPath("/api/users/admin123"); !assert.Equal(3, r) {
		return
	}
	if r := currentNode.LookupPath("/api/users/admin/"); !assert.Equal(3, r) {
		return
	}

	// prefix has trailing slash, but value not defined
	if r := currentNode.LookupPath("/last-root"); !assert.Equal(0, r) {
		return
	}
	if r := currentNode.LookupPath("/last-root/not-found"); !assert.Equal(0, r) {
		return
	}
}

func TestRadix_Remove(t *testing.T) {
	t.Run("base", func(t *testing.T) {
		assert := assert.New(t)

		// r
		// ├── om (8)
		// │   ├── ulus (3)
		// │   └── an
		// │       ├── e (1)
		// │       └── us (2)
		// └── ub
		//     ├── e
		//     │   ├── ns (4)
		//     │   └── r (5)
		//     └── ic
		//         ├── on (6)
		//         └── undus (7)
		currentNode := new(Radix)
		currentNode.Insert("romane", 1)
		currentNode.Insert("romanus", 2)
		currentNode.Insert("romulus", 3)
		currentNode.Insert("rubens", 4)
		currentNode.Insert("ruber", 5)
		currentNode.Insert("rubicon", 6)
		currentNode.Insert("rubicundus", 7)
		currentNode.Insert("rom", 8)

		e := currentNode.edge("r").node.edge("om")
		assert.Len(e.node.edges, 2)

		// remove empty node
		if v, ok := currentNode.Remove("romulus"); assert.True(ok) {
			assert.Equal(3, v)
		}
		assert.Len(e.node.edges, 1)

		// "an" and "us" should be merged into "anus" :)
		if v, ok := currentNode.Remove("romane"); assert.True(ok) {
			assert.Equal(1, v)
		}
		assert.Equal("anus", e.node.edges[0].prefix)

		// return "an" back
		currentNode.Insert("roman", 9)
		// remove node with child
		if v, ok := currentNode.Remove("rom"); assert.True(ok) {
			assert.Equal(8, v)
		}
		// "om" should be merged with "an"
		e = currentNode.edge("r").node.edge("oman")
		assert.Len(e.node.edges, 1)
		assert.Equal(9, e.node.value)
		e = e.node.edge("us")
		assert.Len(e.node.edges, 0)
		assert.Equal(2, e.node.value)

		currentNode.Dump(os.Stdout)
	})

	t.Run("remove twice", func(t *testing.T) {
		assert := assert.New(t)

		currentNode := new(Radix)
		currentNode.Insert("apple", 1)
		currentNode.Insert("orange", 2)

		assert.Len(currentNode.edges, 2)

		if v, ok := currentNode.Remove("orange"); assert.True(ok) {
			assert.Equal(2, v)
		}
		assert.Len(currentNode.edges, 1)

		_, ok := currentNode.Remove("orange")
		assert.False(ok)
		assert.Len(currentNode.edges, 1)
	})
}

func BenchmarkInsert(b *testing.B) {
	r := new(Radix)

	var keys []string
	for n := 0; n < b.N; n++ {
		key := strconv.Itoa(n)
		keys = append(keys, key)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _ = r.Insert(keys[n], true)
	}
}

func BenchmarkInsertRemove(b *testing.B) {
	r := new(Radix)

	var keys []string
	for n := 0; n < b.N; n++ {
		key := strconv.Itoa(n)
		keys = append(keys, key)
		_, _ = r.Insert(key, true)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _ = r.Remove(keys[n])
	}
}

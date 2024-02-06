package shardmap_test

import (
	"testing"

	"github.com/saweima12/gcrate/shardmap"
)

type TestKeyItem struct {
	Key string
}

func (te *TestKeyItem) String() string {
	return te.Key
}

type TestValueItem struct {
	Value string
}

func TestShardMap(t *testing.T) {
	nm := shardmap.NewStringer[*TestKeyItem, *TestValueItem]()

	nm.Set(&TestKeyItem{Key: "100"}, &TestValueItem{Value: "Woo"})
	nm.Set(&TestKeyItem{Key: "300"}, &TestValueItem{Value: "Woo"})

}

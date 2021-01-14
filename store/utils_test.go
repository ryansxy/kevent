package store

import (
	"github.com/k0kubun/pp"
	"testing"
)

func TestFind(t *testing.T) {
	rm := RequestMessage{Clustername: "cluster1", Kind: "Deployment"}
	e, _ := Find(rm.Clustername, rm)
	pp.Println(e)
}

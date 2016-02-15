package discovery

import (
	"testing"

	bh "github.com/kandoo/beehive"
	"github.com/kandoo/beehive-netctrl/nom"
)

func TestGraphBuilderCentralizedSinglePath(t *testing.T) {
	links := []nom.Link{
		{From: "n2$$1", To: "n1$$1"},
		{From: "n6$$2", To: "n5$$1"},
		{From: "n5$$2", To: "n1$$1"},
		{From: "n5$$2", To: "n6$$1"},
		{From: "n7$$2", To: "n5$$2"},
		{From: "n4$$3", To: "n2$$1"},
		{From: "n3$$3", To: "n2$$2"},
		{From: "n1$$1", To: "n2$$3"},
		{From: "n1$$2", To: "n5$$3"},
		{From: "n5$$3", To: "n7$$2"},
		{From: "n2$$1", To: "n4$$3"},
		{From: "n2$$2", To: "n3$$3"},
	}
	b := GraphBuilderCentralized{}
	ctx := &bh.MockRcvContext{}
	for _, l := range links {
		msg := &bh.MockMsg{
			MsgData: nom.LinkAdded(l),
		}
		b.Rcv(msg, ctx)
	}
	paths, l := ShortestPathCentralized("n3", "n1", ctx)
	if l != 2 {
		t.Errorf("invalid shortest path between n1 and n6: actual=%d want=2", l)
	}
	if len(paths) != 1 {
		t.Errorf("invalid number of paths between n1 and n6: actual=%d want=1",
			len(paths))
	}
	for _, p := range paths {
		if p[1] != links[2] && p[1] != links[3] {
			t.Errorf("invalid path: %v", p)
		}
	}
}

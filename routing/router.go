package routing

import (

	"fmt"
	"time"

	bh "github.com/kandoo/beehive"
    "github.com/kandoo/beehive-netctrl/nom"
    "github.com/jyzhe/beehive-netctrl/discovery"
	"github.com/jyzhe/beehive-netctrl/switching"

)

const (
	Mac2uidDict = "mac2uid"
)

// InstallRouting installs the routing application on bh.DefaultHive.
// timeout is the duration between each epoc of routing advertisements.
func InstallRouting(timeout time.Duration, h bh.Hive, opts ...bh.AppOption) {

    app := h.NewApp("Routing", opts...)
	router := Router{}
    app.Handle(nom.PacketIn{}, router)


	builder := discovery.GraphBuilderCentralized{}
	app.Handle(nom.LinkAdded{}, builder)
	app.Handle(nom.LinkDeleted{}, builder)

	app.Handle(nom.NodeJoined{}, nodeJoinedHandler2{})
	app.Handle(nom.NodeLeft{}, nodeLeftHandler2{})

    fmt.Println("Installing Router....")
}

// Router is the main handler of the routing application.
type Router struct{
	switching.Hub
}

// Rcv handles both Discovery and Advertisement messages.
func (r Router) Rcv(msg bh.Msg, ctx bh.RcvContext) error {

    in := msg.Data().(nom.PacketIn)
	src := in.Packet.SrcMAC()
	dst := in.Packet.DstMAC()

	if dst.IsLLDP() {
		fmt.Println("Router Rcv: Received LLDP")
		return nil
	}

	if dst.IsBroadcast() || dst.IsMulticast() {
		fmt.Println("Router Rcv: Received Broadcast or Multicast")
		return r.Hub.Rcv(msg, ctx)
	}

	fmt.Println("Router Rcv: Received packet in....")
	d := ctx.Dict(Mac2uidDict)
	srck := src.Key()
	sv, err := d.Get(srck);
	if err != nil {
		fmt.Printf("Router Rcv: Cant find source node %v\n", srck)
		return nil
	}
	s_id := sv.(nom.UID)

	dstk := dst.Key()
	dv, err := d.Get(dstk);
	if  err != nil {
		fmt.Printf("Router Rcv: Cant find dest node %v\n", dstk)
		return nil
	}
	d_id := dv.(nom.UID)

	paths, _ := discovery.ShortestPathCentralized(s_id, d_id, ctx)

	fmt.Printf("Router Rcv: Path returns %v\n", paths)
    return nil

}

// Rcv maps Discovery based on its destination node and Advertisement messages
// based on their source node.
func (r Router) Map(msg bh.Msg, ctx bh.MapContext) bh.MappedCells {

	return bh.MappedCells{{"N", string(msg.Data().(nom.PacketIn).Node)}}

}


type nodeJoinedHandler2 struct {}

func (h nodeJoinedHandler2) Rcv(msg bh.Msg, ctx bh.RcvContext) error {

	joined := msg.Data().(nom.NodeJoined)
	d := ctx.Dict(Mac2uidDict)
	n := nom.Node(joined)
	k := n.MACAddr.Key()
	fmt.Printf("Router: Adding %v to dict\n", k)
	if _, err := d.Get(k); err == nil {
		return nil
	}

	return d.Put(k, string(n.UID()))

}

func (h nodeJoinedHandler2) Map(msg bh.Msg, ctx bh.MapContext) bh.MappedCells {

	joined := msg.Data().(nom.NodeJoined)
	n := nom.Node(joined)

	return bh.MappedCells{{"N", string(n.UID())}}
}

type nodeLeftHandler2 struct {}

func (h nodeLeftHandler2) Rcv(msg bh.Msg, ctx bh.RcvContext) error {
	left := msg.Data().(nom.NodeLeft)
	d := ctx.Dict(Mac2uidDict)
	n := nom.Node(left)
	k := n.MACAddr.Key()
	if _, err := d.Get(k); err != nil {
		return fmt.Errorf("%v is not joined", n)
	}
	d.Del(k)
	return nil

}

func (h nodeLeftHandler2) Map(msg bh.Msg, ctx bh.MapContext) bh.MappedCells {

	left := msg.Data().(nom.NodeLeft)
	n := nom.Node(left)

	return bh.MappedCells{{"N", string(n.UID())}}
}

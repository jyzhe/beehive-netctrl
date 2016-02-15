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
	mac2port = "mac2port"
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

	d := ctx.Dict(mac2port)

	if dst.IsLLDP() {
		// fmt.Printf("Router Rcv: Received LLDP from %v to %v\n", src, dst)
		return nil
	}

	// Save the endhost info
	srck := src.Key()
	if _, get_err := d.Get(srck); get_err != nil {
		fmt.Printf("Router Rcv: Saving %v, %v to dict (%v)\n", srck, in.InPort, get_err)
		if put_err := d.Put(srck, in.InPort); put_err != nil {
			fmt.Println("****Router Rcv: Save source error!")
		}
	}

	if dst.IsBroadcast() || dst.IsMulticast() {
		// fmt.Printf("Router Rcv: Received Broadcast or Multicast from %v to %v, innode is %v, %v\n", src, dst, in.Node, in.InPort)
		return r.Hub.Rcv(msg, ctx)
	}

	// fmt.Printf("Router Rcv: Received packet in from %v to %v, innode is %v, %v\n", src, dst, in.Node, in.InPort)

	sn := in.Node

	dstk := dst.Key()
	dv, err := d.Get(dstk)
	if  err != nil {
		fmt.Printf("Router Rcv: Cant find dest node %v\n", dstk)
		return nil
	}
	dn,_ := nom.ParsePortUID(dv.(nom.UID))

	paths, len := discovery.ShortestPathCentralized(sn, nom.UID(dn), ctx)

	// fmt.Printf("Router Rcv: Path between %v and %v returns %v, %v\n", sn, nom.UID(dn), paths, len)

	p := dv.(nom.UID)
	if len > 0 {
		shortest := paths[0]
		p = shortest[0].From

		// add reverse flow entry
		add := nom.AddFlowEntry{
			Flow: nom.FlowEntry{
				Node: in.Node,
				Match: nom.Match{
					Fields: []nom.Field{
						nom.EthDst{
							Addr: src,
							Mask: nom.MaskNoneMAC,
						},
					},
				},
				Actions: []nom.Action{
					nom.ActionForward{
						Ports: []nom.UID{in.InPort},
					},
				},
			},
		}
		ctx.Reply(msg, add)
	}

	out := nom.PacketOut{
		Node:     in.Node,
		InPort:   in.InPort,
		BufferID: in.BufferID,
		Packet:   in.Packet,
		Actions: []nom.Action{
			nom.ActionForward{
				Ports: []nom.UID{p},
			},
		},
	}
	ctx.Reply(msg, out)

    return nil

}

// Rcv maps Discovery based on its destination node and Advertisement messages
// based on their source node.
func (r Router) Map(msg bh.Msg, ctx bh.MapContext) bh.MappedCells {

	// return bh.MappedCells{{"N", string(msg.Data().(nom.PacketIn).Node)}}
	return bh.MappedCells{{"__D__", "__0__"}}

}

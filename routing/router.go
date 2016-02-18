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

	// builder := discovery.GraphBuilderCentralized{}
	app.Handle(nom.LinkAdded{}, router)
	app.Handle(nom.LinkDeleted{}, router)

    fmt.Println("Installing Router....")
}

// Router is the main handler of the routing application.
type Router struct{
	switching.Hub
	discovery.GraphBuilderCentralized
}

func registerEndhosts(ctx bh.RcvContext) error {
	d := ctx.Dict(mac2port)
	a1 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	d.Put(nom.MACAddr(a1).Key(), nom.UID("5$$1"))
	a2 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02}
	d.Put(nom.MACAddr(a2).Key(), nom.UID("5$$2"))
	a3 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x03}
	d.Put(nom.MACAddr(a3).Key(), nom.UID("6$$1"))
	a4 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x04}
	d.Put(nom.MACAddr(a4).Key(), nom.UID("6$$2"))
	a5 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x05}
	d.Put(nom.MACAddr(a5).Key(), nom.UID("9$$1"))
	a6 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x06}
	d.Put(nom.MACAddr(a6).Key(), nom.UID("9$$2"))
	a7 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x07}
	d.Put(nom.MACAddr(a7).Key(), nom.UID("a$$1"))
	a8 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x08}
	d.Put(nom.MACAddr(a8).Key(), nom.UID("a$$2"))
	return nil
}

// Rcv handles both Discovery and Advertisement messages.
func (r Router) Rcv(msg bh.Msg, ctx bh.RcvContext) error {

	switch msg.Data().(type) {
	case nom.LinkAdded:
		return r.GraphBuilderCentralized.Rcv(msg, ctx)
	case nom.LinkDeleted:
		return r.GraphBuilderCentralized.Rcv(msg, ctx)
	default:
		in := msg.Data().(nom.PacketIn)
		src := in.Packet.SrcMAC()
		dst := in.Packet.DstMAC()

		d := ctx.Dict(mac2port)

		if dst.IsLLDP() {
			return nil
		}

		// TODO: Maybe there are alternative ways to get device info
		// 		Or devide the network into areas and use masks
		srck := src.Key()
		_, src_err := d.Get(srck)
		if src_err != nil {

			// fmt.Printf("Router Rcv: Register all nodes %v at %v\n", src, in.InPort)
			registerEndhosts(ctx)
			// if put_err := d.Put(srck, in.InPort); put_err != nil {
			// 	fmt.Println("****Router Rcv: Save source error!")
			// }
		}

		if dst.IsBroadcast() || dst.IsMulticast() {
			// fmt.Printf("Router Rcv: Received Broadcast or Multicast from %v to %v, innode is %v, %v\n", src, dst, in.Node, in.InPort)
			return r.Hub.Rcv(msg, ctx)
		}

		sn := in.Node

		// fmt.Printf("Router Rcv: Received packet in from %v to %v, innode is %v, %v\n", src, dst, in.Node, in.InPort)

		dstk := dst.Key()
		dst_port, dst_err := d.Get(dstk)
		if  dst_err != nil {
			fmt.Printf("Router Rcv: Cant find dest node %v\n", dstk)
			return nil
		}
		dn,_ := nom.ParsePortUID(dst_port.(nom.UID))
		p := dst_port.(nom.UID)

		if (sn != nom.UID(dn)){

			paths, shortest_len := discovery.ShortestPathCentralized(sn, nom.UID(dn), ctx)
			fmt.Printf("Router Rcv: Path between %v and %v returns %v, %v\n", sn, nom.UID(dn), paths, shortest_len)

			for _, path := range paths {
				if len(path) != shortest_len {
					continue
				} else {

					p = path[0].From
					if src_err == nil {

						// nf, _ := nom.ParsePortUID(src_port.(nom.UID))
						// if nom.UID(nf) != in.Node {
						add_forward := nom.AddFlowEntry{
							Flow: nom.FlowEntry{
								Node: in.Node,
								Match: nom.Match{
									Fields: []nom.Field{
										nom.EthDst{
											Addr: dst,
											Mask: nom.MaskNoneMAC,
										},
									},
								},
								Actions: []nom.Action{
									nom.ActionForward{
										Ports: []nom.UID{p},
									},
								},
							},
						}
						ctx.Reply(msg, add_forward)
						// }

					}

					if dst_err == nil {

						// nt, _ := nom.ParsePortUID(dst_port.(nom.UID))
						// if nom.UID(nt) != in.Node {
						add_reverse := nom.AddFlowEntry{
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
						ctx.Reply(msg, add_reverse)

						// }
					}
					break
				}
			}
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

		// fmt.Printf("Router Rcv: Received packet in from %v,%v to %v,%v\n", src, in.InPort, dst, p)

	}

    return nil

}

// Rcv maps Discovery based on its destination node and Advertisement messages
// based on their source node.
func (r Router) Map(msg bh.Msg, ctx bh.MapContext) bh.MappedCells {

	return bh.MappedCells{{"__D__", "__0__"}}

	// switch dm := msg.Data().(type) {
	// case nom.LinkAdded:
	// 	from := dm.From
	// 	n, _ := nom.ParsePortUID(from)
	// 	return bh.MappedCells{{"N", string(n)}}
	// 	// return bh.MappedCells{{"__D__", "__0__"}}
	// case nom.LinkDeleted:
	// 	from := dm.From
	// 	n, _ := nom.ParsePortUID(from)
	// 	return bh.MappedCells{{"N", string(n)}}
	// 	// return bh.MappedCells{{"__D__", "__0__"}}
	// default:
	// 	return bh.MappedCells{{"N", string(msg.Data().(nom.PacketIn).Node)}}
	// 	// return bh.MappedCells{{"__D__", "__0__"}}
	// }
}

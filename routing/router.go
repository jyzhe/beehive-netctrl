package routing

import (
	"fmt"

	"github.com/jyzhe/beehive-netctrl/discovery"
	bh "github.com/kandoo/beehive"
	"github.com/kandoo/beehive-netctrl/nom"
	"github.com/kandoo/beehive/Godeps/_workspace/src/golang.org/x/net/context"
)

// Router is the main handler of the routing application.
type Router struct {
	// switching.Hub
	discovery.GraphBuilderCentralized
}

func (r Router) Rcv(msg bh.Msg, ctx bh.RcvContext) error {

	switch dm := msg.Data().(type) {
	case setup:
		return registerEndhosts(ctx)
	case nom.LinkAdded:
		link := InterAreaLink(dm)
		ctx.Emit(link)
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

		// FIXME: Hardcoding the hardware address at the moment
		srck := src.Key()
		_, src_err := d.Get(srck)
		if src_err != nil {
			fmt.Printf("Router: Error retrieving hosts %v\n", src)
		}

		if dst.IsBroadcast() || dst.IsMulticast() {
			fmt.Printf("Router: Received Broadcast or Multicast from %v\n", src)
			return nil
		}

		sn := in.Node

		dstk := dst.Key()
		dst_port, dst_err := d.Get(dstk)
		if dst_err != nil {
			fmt.Printf("Router: Cant find dest node %v\n", dstk)
			res, query_err := ctx.Sync(context.TODO(), InterAreaQuery{Src: srck, Dst: dstk})
			if query_err != nil {
				fmt.Printf("Router: received error when querying! %v\n", query_err)
			}
			fmt.Printf("Router: received response succesfully - %v\n", res)
			dst_port = res.(nom.UID)
		}
		dn, _ := nom.ParsePortUID(dst_port.(nom.UID))
		p := dst_port.(nom.UID)

		if sn != nom.UID(dn) {

			paths, shortest_len := discovery.ShortestPathCentralized(sn, nom.UID(dn), ctx)
			fmt.Printf("Router: Path between %v and %v returns %v, %v\n", sn, nom.UID(dn), paths, shortest_len)

			for _, path := range paths {
				if len(path) != shortest_len {
					continue
				} else {

					p = path[0].From
					break
				}
			}
		}

		// Forward flow entry
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

		// Reverse flow entry
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
	// case nom.LinkDeleted:
	// 	from := dm.From
	// 	n, _ := nom.ParsePortUID(from)
	// 	return bh.MappedCells{{"N", string(n)}}
	// default:
	// 	return bh.MappedCells{{"N", string(msg.Data().(nom.PacketIn).Node)}}
	// }
}

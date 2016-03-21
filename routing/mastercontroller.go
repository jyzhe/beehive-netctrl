package routing

import (

	"fmt"

	bh "github.com/kandoo/beehive"
    "github.com/jyzhe/beehive-netctrl/discovery"
    // "github.com/kandoo/beehive-netctrl/nom"
)

// Router is the main handler of the routing application.
type MasterController struct{
	discovery.GraphBuilderCentralized
}

func (m MasterController) Rcv(msg bh.Msg, ctx bh.RcvContext) error {

	switch msg.Data().(type) {
	case area_setup:
		return master_setup(ctx)

	case InterAreaLink:
        fmt.Println("Received InterAreaLink!")
		return nil
    case InterAreaQuery:
        fmt.Println("Received Inter Area Query")
        return nil
	// case nom.LinkDeleted:
	// 	return r.GraphBuilderCentralized.Rcv(msg, ctx)
	// default:
	// 	in := msg.Data().(nom.PacketIn)
	// 	src := in.Packet.SrcMAC()
	// 	dst := in.Packet.DstMAC()
    //
	// 	d := ctx.Dict(mac2port)
    //
	// 	if dst.IsLLDP() {
	// 		return nil
	// 	}
    //
	// 	// FIXME: Hardcoding the hardware address at the moment
	// 	srck := src.Key()
	// 	_, src_err := d.Get(srck)
	// 	if src_err != nil {
	// 		fmt.Printf("Router: Error retrieving hosts %v\n", src)
	// 	}
    //
	// 	if dst.IsBroadcast() || dst.IsMulticast() {
	// 		fmt.Printf("Router: Received Broadcast or Multicast from %v\n", src)
	// 		return nil
	// 	}
    //
	// 	sn := in.Node
    //
	// 	dstk := dst.Key()
	// 	dst_port, dst_err := d.Get(dstk)
	// 	if  dst_err != nil {
	// 		fmt.Printf("Router: Cant find dest node %v\n", dstk)
	// 		dst_port, _ = d.Get("default")
	// 	}
	// 	dn,_ := nom.ParsePortUID(dst_port.(nom.UID))
	// 	p := dst_port.(nom.UID)
    //
	// 	if (sn != nom.UID(dn)){
    //
	// 		paths, shortest_len := discovery.ShortestPathCentralized(sn, nom.UID(dn), ctx)
	// 		fmt.Printf("Router: Path between %v and %v returns %v, %v\n", sn, nom.UID(dn), paths, shortest_len)
    //
	// 		// if shortest_len == -1 {
	// 		// 	borderDict := ctx.Dict(border_node_dict)
	// 		// 	borderDict.ForEach(func(k string, v interface{}) bool{
	// 		// 		node, _ := nom.ParsePortUID(v.(nom.UID))
	// 		// 		paths, shortest_len = discovery.ShortestPathCentralized(sn, nom.UID(node), ctx)
	// 		//
	// 		// 		if shortest_len == 0{
	// 		// 			p = v.(nom.UID)
	// 		// 		}
	// 		// 		// TODO: Find the shortest one
	// 		// 		return false
	// 		// 	})
	// 		// }
	// 		fmt.Printf("Router: After adjustment length: %v\n", shortest_len)
	// 		for _, path := range paths {
	// 			if len(path) != shortest_len {
	// 				continue
	// 			} else {
    //
	// 				p = path[0].From
	// 				break
	// 			}
	// 		}
	// 	}
    //
	// 	if src_err == nil {
    //
	// 		// Forward flow entry
	// 		add_forward := nom.AddFlowEntry{
	// 			Flow: nom.FlowEntry{
	// 				Node: in.Node,
	// 				Match: nom.Match{
	// 					Fields: []nom.Field{
	// 						nom.EthDst{
	// 							Addr: dst,
	// 							Mask: nom.MaskNoneMAC,
	// 						},
	// 					},
	// 				},
	// 				Actions: []nom.Action{
	// 					nom.ActionForward{
	// 						Ports: []nom.UID{p},
	// 					},
	// 				},
	// 			},
	// 		}
	// 		ctx.Reply(msg, add_forward)
    //
	// 	}
    //
	// 	if dst_err == nil {
    //
	// 		// Reverse flow entry
	// 		add_reverse := nom.AddFlowEntry{
	// 			Flow: nom.FlowEntry{
	// 				Node: in.Node,
	// 				Match: nom.Match{
	// 					Fields: []nom.Field{
	// 						nom.EthDst{
	// 							Addr: src,
	// 							Mask: nom.MaskNoneMAC,
	// 						},
	// 					},
	// 				},
	// 				Actions: []nom.Action{
	// 					nom.ActionForward{
	// 						Ports: []nom.UID{in.InPort},
	// 					},
	// 				},
	// 			},
	// 		}
	// 		ctx.Reply(msg, add_reverse)
    //
	// 	}
    //
	// 	out := nom.PacketOut{
	// 		Node:     in.Node,
	// 		InPort:   in.InPort,
	// 		BufferID: in.BufferID,
	// 		Packet:   in.Packet,
	// 		Actions: []nom.Action{
	// 			nom.ActionForward{
	// 				Ports: []nom.UID{p},
	// 			},
	// 		},
	// 	}
	// 	ctx.Reply(msg, out)
	}

    return nil

}

// Rcv maps Discovery based on its destination node and Advertisement messages
// based on their source node.
func (m MasterController) Map(msg bh.Msg, ctx bh.MapContext) bh.MappedCells {
	return bh.MappedCells{{"__D__", "__0__"}}
}

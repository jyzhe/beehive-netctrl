package routing

import (

	"fmt"

	bh "github.com/kandoo/beehive"
    "github.com/kandoo/beehive-netctrl/nom"
    "github.com/jyzhe/beehive-netctrl/discovery"
)

// Router is the main handler of the routing application.
type LoadBalancer struct{

	discovery.GraphBuilderCentralized
}

func (r LoadBalancer) Rcv(msg bh.Msg, ctx bh.RcvContext) error {

	switch msg.Data().(type) {
    case setup:
        return routing_setup(ctx)
	case nom.LinkAdded:
		return r.GraphBuilderCentralized.Rcv(msg, ctx)
	case nom.LinkDeleted:
		return r.GraphBuilderCentralized.Rcv(msg, ctx)
	default:
		in := msg.Data().(nom.PacketIn)
		src := in.Packet.SrcMAC()
		dst := in.Packet.DstMAC()

		d := ctx.Dict(mac2port)
        load_dict := ctx.Dict(load_on_nodes)

		if dst.IsLLDP() {
			return nil
		}

		// FIXME: Hardcoding the hardware address at the moment
		srck := src.Key()
		_, src_err := d.Get(srck)
		if src_err != nil {
			registerEndhosts(ctx)
		}

		if dst.IsBroadcast() || dst.IsMulticast() {
			fmt.Printf("Load Balancer: Received Broadcast or Multicast from %v\n", src)
			return nil
		}

		sn := in.Node

		dstk := dst.Key()
		dst_port, dst_err := d.Get(dstk)
		if  dst_err != nil {
			fmt.Printf("Load Balancer: Cant find dest node %v\n", dstk)
			return nil
		}
		dn,_ := nom.ParsePortUID(dst_port.(nom.UID))
		p := dst_port.(nom.UID)

		if (sn != nom.UID(dn)){

			paths, _ := discovery.ShortestPathCentralized(sn, nom.UID(dn), ctx)

            opt_path := paths[0]
            min_load := calculate_load(ctx, paths[0])
			for _, path := range paths[1:] {

                load := calculate_load(ctx, path)
                if load < min_load{
                    opt_path = path
                    min_load = load
                }
			}

            fmt.Printf("Load Balancer: Path between %v and %v returns %v, %v\n", sn, nom.UID(dn), opt_path, len(opt_path))
            p = opt_path[0].From
		}

        if src_err == nil {

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

        }

        if dst_err == nil {

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

        }

        // Updating the load on this node
        // FIXME: This is a naive approach, ideally should update when flowentry
        // is added/removed, but flowentry deleted is not implemented yet
        if load, err := load_dict.Get(string(in.Node)); err == nil {
            load_dict.Put(string(in.Node), load.(int) + 1)
        } else {
            load_dict.Put(string(in.Node), 1)
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
	}

    return nil

}

// Rcv maps Discovery based on its destination node and Advertisement messages
// based on their source node.
func (r LoadBalancer) Map(msg bh.Msg, ctx bh.MapContext) bh.MappedCells {

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

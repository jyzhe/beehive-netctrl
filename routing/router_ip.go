package routing

import (

    "fmt"

    bh "github.com/jyzhe/beehive"
    "github.com/jyzhe/beehive-netctrl/nom"
    "github.com/jyzhe/beehive-netctrl/discovery"
)

// Router is the main handler of the routing application.
type RouterIP struct{
    // switching.Hub
    discovery.GraphBuilderCentralized
}

func (r RouterIP) Rcv(msg bh.Msg, ctx bh.RcvContext) error {

    switch msg.Data().(type) {
    case setupIP:
        return routingIP_setup(ctx)
    case nom.LinkAdded:
        return r.GraphBuilderCentralized.Rcv(msg, ctx)
    case nom.LinkDeleted:
        return r.GraphBuilderCentralized.Rcv(msg, ctx)
    default:
        in := msg.Data().(nom.PacketIn)
        src := in.Packet.SrcMAC()
        dst := in.Packet.DstMAC()
        src_ip := in.Packet.SrcIP()
        dst_ip := in.Packet.DstIP()
        d := ctx.Dict(ip2port)

        if dst.IsLLDP() {
            return nil
        }

        // FIXME: Hardcoding the hardware address at the moment
        srck := src_ip.Key()
        _, src_err := d.Get(srck)
        if src_err != nil {
            fmt.Printf("Router: Error retrieving hosts %v\n", src)
        }

        if dst.IsBroadcast() || dst.IsMulticast() {
            fmt.Printf("Router: Received Broadcast or Multicast from %v\n", src)
            return nil
        }

        sn := in.Node

        dstk := dst_ip.Key()
        dst_port, dst_err := d.Get(dstk)
        if  dst_err != nil {
            fmt.Printf("Router: Cant find dest node %v\n", dstk)
            return nil
        }
        dn,_ := nom.ParsePortUID(dst_port.(nom.UID))
        p := dst_port.(nom.UID)

        if (sn != nom.UID(dn)){

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

        if src_err == nil {

            // Forward flow entry
            add_forward := nom.AddFlowEntry{
                Flow: nom.FlowEntry{
                    Node: in.Node,
                    Match: nom.Match{
                        Fields: []nom.Field{
                            nom.IPv4Dst{
                                Addr: dst_ip,
                                Mask: nom.MaskNoneIPV4,
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
                            nom.IPv4Src{
                                Addr: src_ip,
                                Mask: nom.MaskNoneIPV4,
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
func (r RouterIP) Map(msg bh.Msg, ctx bh.MapContext) bh.MappedCells {

    return bh.MappedCells{{"__D__", "__0__"}}

    // switch dm := msg.Data().(type) {
    // case nom.LinkAdded:
    //  from := dm.From
    //  n, _ := nom.ParsePortUID(from)
    //  return bh.MappedCells{{"N", string(n)}}
    // case nom.LinkDeleted:
    //  from := dm.From
    //  n, _ := nom.ParsePortUID(from)
    //  return bh.MappedCells{{"N", string(n)}}
    // default:
    //  return bh.MappedCells{{"N", string(msg.Data().(nom.PacketIn).Node)}}
    // }
}

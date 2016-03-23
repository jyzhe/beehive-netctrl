package routing

import (

    "fmt"

    bh "github.com/kandoo/beehive"
    "github.com/kandoo/beehive-netctrl/nom"
    "github.com/jyzhe/beehive-netctrl/discovery"
    "github.com/kandoo/beehive/Godeps/_workspace/src/golang.org/x/net/context"
)

// Router is the main handler of the routing application.
type RouterIP struct{
    // switching.Hub
    discovery.GraphBuilderCentralized
}

type areaQuery struct{
    dst_ip nom.IPv4Addr
    src_ip nom.IPv4Addr
}

func FindAreaId(dst_ip nom.IPv4Addr) string{
    return string(dst_ip[0])
}

func (r RouterIP) Rcv(msg bh.Msg, ctx bh.RcvContext) error {

    switch msg.Data().(type) {
    case nom.NodeJoined:
        ctx.Printf("node joined\n")
    case setup:
        return routing_setupIP(ctx)
    case nom.LinkAdded:
        return r.GraphBuilderCentralized.Rcv(msg, ctx)
    case nom.LinkDeleted:
        return r.GraphBuilderCentralized.Rcv(msg, ctx)
    case areaQuery:
        dst_area_id := FindAreaId((msg.Data().(areaQuery).dst_ip))
        src_area_id := FindAreaId((msg.Data().(areaQuery).src_ip))
        ctx.Printf("area %s to area %s",src_area_id,dst_area_id)
        //TO DO THIS IS FINDING THE SHORTEST PATH IN SHORTEST GRAPH
        // dst_border_nodes = ctx.Dict(border_dict).get(dst_area_id)
        // src_border_nodes = ctx.Dict(border_dict).get(src_area_id)
        // for _, dst_node := range dst_border_nodes{
        //     for _, src_node := range src_border_nodes{
        //         paths, shortest_len := discovery.ShortestPathCentralized(src_node, dst_node, ctx)
        //     }
        // }
        // for _, path := range paths {
        //     if len(path) >= shortest_len {
        //         continue
        //     } 
        //     else {
        //         shortest_p = path[0].From
        //     }
        // }
        // ctx.ReplyTo(msg, shortest_p)

    case nom.PacketIn:
        in := msg.Data().(nom.PacketIn)
        // src := in.Packet.SrcMAC()
        dst := in.Packet.DstMAC()
        src_ip := SrcIP(in.Packet)
        dst_ip := DstIP(in.Packet)
        fmt.Printf("src ip:%s, dst ip:%s\n",src_ip.String(),dst_ip.String())

        ip2portdict := ctx.Dict(ip2port)
        areaId:=FindAreaId(src_ip)
        ctx.Printf("%s\n",areaId)
        d,_ := ip2portdict.Get(areaId)
        if d == nil{
            fmt.Printf("Dictionary not existed\n")
        }
        if dst.IsLLDP() {
            return nil
        }

        // FIXME: Hardcoding the hardware address at the moment
        srck := src_ip.String()
        _, ok := d.(map[string]nom.UID)[srck]
        if ok != true {
            fmt.Printf("Router: Error retrieving hosts %v\n", src_ip)
        }

        if dst.IsBroadcast() || dst.IsMulticast() {
            fmt.Printf("Router: Received Broadcast or Multicast from %v\n", src_ip)
            return nil
        }

        sn := in.Node

        dstk := dst_ip.String()
        dst_port, dst_err := d.(map[string]nom.UID)[dstk]
        if  dst_err != true {
            fmt.Printf("Router: Cant find dest node %v\n", dstk)
            _, reply_err := bh.Sync(context.TODO(), areaQuery{dst_ip, src_ip})
            if reply_err!= nil{
                fmt.Printf("No return messge for the query\n")
                return nil
            }


            return nil
        }
        dn,_ := nom.ParsePortUID(dst_port)
        p := dst_port

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
        //FLOW entry comment:
        // if src_err == nil {

        //     // Forward flow entry
        //     add_forward := nom.AddFlowEntry{
        //         Flow: nom.FlowEntry{
        //             Node: in.Node,
        //             Match: nom.Match{
        //                 Fields: []nom.Field{
        //                     nom.IPv4Dst{
        //                         Addr: dst_ip,
        //                         Mask: nom.MaskNoneIPV4,
        //                     },
        //                 },
        //             },
        //             Actions: []nom.Action{
        //                 nom.ActionForward{
        //                     Ports: []nom.UID{p},
        //                 },
        //             },
        //         },
        //     }
        //     ctx.Reply(msg, add_forward)
        //     fmt.Println("Add Forward",add_forward);

        // }

        // if dst_err == nil {

        //     // Reverse flow entry
        //     add_reverse := nom.AddFlowEntry{
        //         Flow: nom.FlowEntry{
        //             Node: in.Node,
        //             Match: nom.Match{
        //                 Fields: []nom.Field{
        //                     nom.IPv4Dst{
        //                         Addr: src_ip,
        //                         Mask: nom.MaskNoneIPV4,
        //                     },
        //                 },
        //             },
        //             Actions: []nom.Action{
        //                 nom.ActionForward{
        //                     Ports: []nom.UID{in.InPort},
        //                 },
        //             },
        //         },
        //     }
        //     ctx.Reply(msg, add_reverse)
        //     fmt.Println("Add Reverse",add_reverse);

        // }

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
    switch msg.Data().(type){
    case areaQuery:
        return bh.MappedCells{{"__D__", "__0__"}}
    case nom.PacketIn:
        in := msg.Data().(nom.PacketIn)
        src_ip := SrcIP(in.Packet)
        return bh.MappedCells{{ip2port, FindAreaId(src_ip)}}
    case nom.LinkAdded:
        return bh.MappedCells{{}}
    case nom.LinkDeleted:
        return bh.MappedCells{{}}
    default:
        return bh.MappedCells{{}}
    }

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


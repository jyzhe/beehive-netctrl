package routing

import (

    "fmt"

    bh "github.com/kandoo/beehive"
    "github.com/kandoo/beehive-netctrl/nom"
    "github.com/jyzhe/beehive-netctrl/discovery"
)

// Router is the main handler of the routing application.
type RouterMAC struct{
    // switching.Hub
    discovery.GraphBuilderCentralized
}
const (
    areaTonode = 'areaTonode' // AreaID mapping to a node map!, areaID -> map[node's mac] port UID
    borderNode = 'bordernode' // AreaID mapping to a list of node!, areaID -> map list of node


)
type areaQuery{ // Area query, only give the master conroller this  areaid -> another areaid
    from string
    to string
}

type setupArea{
    area string
}

func FindAreaId(ID nom.NodeID) string{ // This should take a node id, then return an area id. like 1 -> 1, a -> 2. return as string. Consider hardcoded.
    if ID == '1' || ID == '2' || ID == '3' || ID == '4' || ID == '5'{
        return '1'
    }
    else{
        return '2'
    }
}
func FindAreaIdMac(m nom.MACAddr)string{
    a1 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
    a2 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02}
    a3 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x03}
    a4 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x04}
    a5 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x05}
    if m == a1 || m == a2 || m == a3 || m == a4 || m == a5{
        return '1'
    }
    else{
        return '2'
    }
}

func (r Router) Rcv(msg bh.Msg, ctx bh.RcvContext) error {

    switch msg.Data().(type) {
    case areaQuery:
        fmt.Println("HERE")
        return nil
    case setupArea:
        return registerEndhostsMAC(ctx,)
    case nom.LinkAdded:
        return r.GraphBuilderCentralized.Rcv(msg, ctx)
    case nom.LinkDeleted:
        return r.GraphBuilderCentralized.Rcv(msg, ctx)
    default:
        in := msg.Data().(nom.PacketIn)
        src := in.Packet.SrcMAC()
        dst := in.Packet.DstMAC()

        areaTonodeDic := ctx.Dict(areaTonode)
        d := areaTonode.Get(in.Node)
        if dst.IsLLDP() {
            return nil
        }

        // FIXME: Hardcoding the hardware address at the moment
        srck := src.Key()
        _, src_err := d.(map[string]nom.UID)[srck]
        if src_err != true {
            fmt.Printf("Router: Error retrieving hosts %v\n", src)
            return nil
        }

        if dst.IsBroadcast() || dst.IsMulticast() {
            fmt.Printf("Router: Received Broadcast or Multicast from %v\n", src)
            return nil
        }

        sn := in.Node

        dstk := dst.Key()
        dst_port, dst_err := d.(map[string]nom.UID)[dstk]
        if  dst_err != true {
            fmt.Printf("Router: Cant find dest node %v\n", dstk)
            ctx.Emit(areaQuery{})
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
    switch msg.Data().(type){
    case setupArea:
        set := msg.Data().(setupArea)
        return bh.MappedCells{{areaTonode, set.area}}
    case nom.PacketIn:
        packetin := msg.Data().(nom.PacketIn)
        ni, _ := nom.ParsePortUID(packetin.Node)
        return bh.MappedCells{{areaTonode, FindAreaId(ni)}}
    case nom.LinkAdded:
        var link = nom.Link
        link = nom.Link(dm)
        nf,_ := nom.ParsePortUID(link.From)
        nt,_ := nom.ParsePortUID(link.From)
        if FindAreaId(nf)!= FindAreaId(nt){
            return bh.MappedCells{{"__D__", "__0__"}}  
        }
        else{
            return bh.MappedCells{{areaTonode, FindAreaId(nf)}}
        }
    case nom.LinkDeleted:
        var link = nom.Link
        link = nom.Link(dm)
        nf,_ := nom.ParsePortUID(link.From)
        nt,_ := nom.ParsePortUID(link.From)
        if FindAreaId(nf)!= FindAreaId(nt){
            return bh.MappedCells{{"__D__", "__0__"}}  
        }
        else{
            return bh.MappedCells{{areaTonode, FindAreaId(nf)}}
        }
    case areaQuery
        return bh.MappedCells{{"__D__", "__0__"}}
    }

    return bh.MappedCells{{"__D__", "__0__"}}
    //"__D__" , "__0__" is the master controller
    //case nom.LinkAdded: {__D__, __0__} , {areaTonode, }

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

func InstallRouterMAC(h bh.Hive, opts ...bh.AppOption) {

    app := h.NewApp("Router", opts...)
    router := RouterMAC{}

    // handle routing of packets
    app.Handle(nom.PacketIn{}, router)
    // building centralized network topology
    app.Handle(nom.LinkAdded{}, router)
    app.Handle(nom.LinkDeleted{}, router)
    app.Handle(areaQuery{}, router)

    app.Handle(setupArea{}, router)
    // h.Emit(setupArea{'2'})
    // app.Handle(discovery.RegisterBorderNode{}, router)

    fmt.Println("Installing Router....")
}

func registerEndhostsMAC(ctx bh.RcvContext, area string){
    gob.Register(map[string]nom.UID{})
    d := ctx.Dict(areaTonode)
    if area == '1'{
        m := make(map[string]nom.UID)
        a1 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
        m[nom.MACAddr(a1).Key()] = nom.UID("4$$1")
        a2 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02}
        m[nom.MACAddr(a2).Key()] = nom.UID("4$$2")
        a3 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x03}
        m[nom.MACAddr(a3).Key()] = nom.UID("5$$1")
        a4 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x04}
        m[nom.MACAddr(a4).Key()] = nom.UID("5$$2")
        d.Put(string(1),m)
    }
    else if area == '2'{
        m := make(map[string]nom.UID)
        a5 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x05}
        m[nom.MACAddr(a5).Key()] = nom.UID("9$$1")
        a6 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x06}
        m[nom.MACAddr(a6).Key()] = nom.UID("9$$2")
        a7 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x07}
        m[nom.MACAddr(a7).Key()] = nom.UID("a$$1")
        a8 := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x08}
        m[nom.MACAddr(a8).Key()] = nom.UID("a$$2") 
        d.Put(string(2),m)
    }


    return nil

}

}

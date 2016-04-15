package routing

import (
    "fmt"
    "strings"

    "github.com/jyzhe/beehive-netctrl/discovery"
    bh "github.com/kandoo/beehive"
    "github.com/kandoo/beehive-netctrl/nom"
    "github.com/kandoo/beehive/Godeps/_workspace/src/golang.org/x/net/context"
)

// Router is the main handler of the routing application.
type RouterM struct {
    discovery.GraphBuilderCentralized
}


func (r RouterM) Rcv(msg bh.Msg, ctx bh.RcvContext) error {

    switch dm := msg.Data().(type) {
    case area_setup:
        fmt.Printf("Adding to ctx: %v\n", ctx)
        d:= ctx.Dict(cDict)
        d.Put("master","master")
        return master_setup(ctx)
    case InterAreaLink:
        link := nom.Link(dm)
        nf, _ := nom.ParsePortUID(link.From)
        k := string(nf)
        nt, _ := nom.ParsePortUID(link.To)
        k2 := string(nt)

        d := ctx.Dict(node2area)
        nf_area, err := d.Get(k)
        nt_area, err2 := d.Get(k2)

        if err != nil || err2 != nil {
            fmt.Printf("Node does not belong to any existing areas! %v, %v, %v\n", err, err2, ctx)
            return nil
        }

        if nf_area != nt_area {
            link.From = nom.UID(nf_area.(string) + "$$" + strings.Replace(string(link.From), "$$", tmpD, 1))
            link.To = nom.UID(nt_area.(string) + "$$" + strings.Replace(string(link.To), "$$", tmpD, 1))
            fmt.Printf("Received Links from different areas, building graphs for %v, %v\n", link.From, link.To)
            ctx.Emit(nom.LinkAdded(link))
        }
        return nil
    case InterAreaQuery:

        query := msg.Data().(InterAreaQuery)
        srck := query.Src
        dstk := query.Dst
        d := ctx.Dict(mac2area)

        src_area, src_err := d.Get(srck)
        dst_area, dst_err := d.Get(dstk)

        if src_err != nil || dst_err != nil {
            fmt.Printf("Error retriving area info: %v, %v\n", src_err, dst_err)
            return nil
        }

        paths, shortest_len := discovery.ShortestPathCentralized(nom.UID(src_area.(string)), nom.UID(dst_area.(string)), ctx)

        if shortest_len <= 0 {
            fmt.Printf("No route exists between area %v and area %v\n", src_area, dst_area)
            return nil
        }

        fmt.Printf("Path between area %v and area %v returned %v\n", src_area, dst_area, paths)

        for _, path := range paths {
            if len(path) != shortest_len {
                continue
            } else {
                _, port := nom.ParsePortUID(path[0].From)
                port_conv := strings.Replace(string(port), tmpD, "$$", 1)
                fmt.Printf("Sending converted port: %v\n", port_conv)
                ctx.Reply(msg, nom.UID(port_conv))
                break
            }
        }

        return nil
    case setupM:
        m := msg.Data().(setupM)
        return registerEndhostsAll(ctx, m.area)
    case nom.LinkAdded:
        // link := InterAreaLink(dm)
        cd := ctx.Dict(cDict)
        _, cerr := cd.Get("master")
        if cerr != nil{
            // ctx.Emit(link)
            return r.GraphBuilderCentralized.Rcv(msg, ctx)
        } else{
            link := nom.Link(dm)
            nf, _ := nom.ParsePortUID(link.From)
            k := string(nf)
            nt, _ := nom.ParsePortUID(link.To)
            k2 := string(nt)

            d := ctx.Dict(node2area)
            nf_area, err := d.Get(k)
            nt_area, err2 := d.Get(k2)

            if err != nil || err2 != nil {
                fmt.Printf("Node does not belong to any existing areas! %v, %v, %v\n", err, err2, ctx)
                return nil
            }

            if nf_area != nt_area {
                link.From = nom.UID(nf_area.(string) + "$$" + strings.Replace(string(link.From), "$$", tmpD, 1))
                link.To = nom.UID(nt_area.(string) + "$$" + strings.Replace(string(link.To), "$$", tmpD, 1))
                fmt.Printf("Received Links from different areas, building graphs for %v, %v\n", link.From, link.To)
                InterAreaLinkAdded(link, ctx)
            }
        }
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
func (m RouterM) Map(msg bh.Msg, ctx bh.MapContext) bh.MappedCells {
    switch dm:=msg.Data().(type){
    case InterAreaQuery:
        return bh.MappedCells{{"__D__", "__0__"}}
    case InterAreaLink:
        return bh.MappedCells{{"__D__", "__0__"}}
    case area_setup:
        return bh.MappedCells{{"__D__", "__0__"}}
    case setupM:
        areaMsg := msg.Data().(setupM)
        return bh.MappedCells{{"area", areaMsg.area}}
    case nom.LinkAdded:
        link := nom.Link(dm)
        nf, _ := nom.ParsePortUID(link.From)
        k := string(nf)
        nt, _ := nom.ParsePortUID(link.To)
        k2 := string(nt)
        nf_area := FindAreaId(k)
        nt_area := FindAreaId(k2)
        if nt_area != nf_area{
            return bh.MappedCells{{"__D__", "__0__"}}
        } else{
            return bh.MappedCells{{"area",nf_area}}
        }
    case nom.PacketIn:
        ni := msg.Data().(nom.PacketIn)
        k := string(ni.Node)
        return bh.MappedCells{{"area", FindAreaId(k)}}
    }
    return bh.MappedCells{{"__D__", "__0__"}}
}

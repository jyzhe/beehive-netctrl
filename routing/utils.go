package routing

import (

    "fmt"

	bh "github.com/kandoo/beehive"
    "github.com/kandoo/beehive-netctrl/nom"

)

const (
	mac2port = "mac2port"
    load_on_nodes = "loadonnodes"
)

type setup struct{}

// InstallRouting installs the routing application
func InstallRouter(h bh.Hive, opts ...bh.AppOption) {

    app := h.NewApp("Router", opts...)
	router := Router{}

    // handle routing of packets
    app.Handle(nom.PacketIn{}, router)
    // building centralized network topology
	app.Handle(nom.LinkAdded{}, router)
	app.Handle(nom.LinkDeleted{}, router)

    app.Handle(setup{}, router)
    h.Emit(setup{})

    fmt.Println("Installing Router....")
}

func InstallLoadBalancer(h bh.Hive, opts ...bh.AppOption) {

    app := h.NewApp("LoadBalancer", opts...)
	loadbalancer := LoadBalancer{}

    // handle routing of packets
    app.Handle(nom.PacketIn{}, loadbalancer)
    // building centralized network topology
	app.Handle(nom.LinkAdded{}, loadbalancer)
	app.Handle(nom.LinkDeleted{}, loadbalancer)

    app.Handle(setup{}, loadbalancer)
    h.Emit(setup{})

    fmt.Println("Installing Load Balancer....")
}

func routing_setup(ctx bh.RcvContext) error {
    return registerEndhosts(ctx)
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

func calculate_load(ctx bh.RcvContext, path []nom.Link) int {

    load_dict := ctx.Dict(load_on_nodes)

    // Each link costs one, each flow entry costs 1
    load := len(path)
    for _, link := range path {
        node_id, _ := nom.ParsePortUID(link.To)
        if l, err := load_dict.Get(string(node_id)); err == nil {
            load = load + l.(int)
        }
    }

    return load

}

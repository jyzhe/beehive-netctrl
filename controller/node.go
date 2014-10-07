package controller

import (
	"fmt"

	"github.com/golang/glog"

	"github.com/soheilhy/beehive-netctrl/nom"
	"github.com/soheilhy/beehive/bh"
)

type nodeConnectedHandler struct{}

func (h *nodeConnectedHandler) Rcv(msg bh.Msg, ctx bh.RcvContext) error {
	nc := msg.Data().(nom.NodeConnected)

	dict := ctx.Dict(nodeDriversDict)
	k := bh.Key(nc.Node.ID)
	n := nodeDrivers{}
	if err := nom.DictGet(dict, k, &n); err != nil {
		n.Node = nc.Node
		n.Drivers = nom.Drivers{nc.Driver}
		if err := dict.PutGob(k, &n); err != nil {
			return err
		}

		glog.V(2).Infof("%v joins", nc.Node)
		ctx.Emit(nom.NodeJoined(nc.Node))
		return nil
	}

	if n.hasDriver(nc.Driver) {
		return fmt.Errorf("Driver %v reconnects to %v", nc.Driver, n.Node)
	}

	n.Drivers = append(n.Drivers, nc.Driver)
	return dict.PutGob(k, &n)
}

func (h *nodeConnectedHandler) Map(msg bh.Msg,
	ctx bh.MapContext) bh.MappedCells {

	nc := msg.Data().(nom.NodeConnected)
	return bh.MappedCells{{nodeDriversDict, bh.Key(nc.Node.ID)}}
}

type nodeDisconnectedHandler struct{}

func (h *nodeDisconnectedHandler) Rcv(msg bh.Msg, ctx bh.RcvContext) error {
	nd := msg.Data().(nom.NodeDisconnected)
	k := bh.Key(nd.Node.ID)
	n := nodeDrivers{}
	if err := ctx.Dict(nodeDriversDict).GetGob(k, &n); err != nil {
		return fmt.Errorf("Driver %v disconnects from %v before connecting",
			nd.Driver, nd.Node)
	}

	if !n.removeDriver(nd.Driver) {
		return fmt.Errorf("Driver %v disconnects from %v before connecting",
			nd.Driver, nd.Node)
	}

	return nil
}

func (h *nodeDisconnectedHandler) Map(msg bh.Msg,
	ctx bh.MapContext) bh.MappedCells {

	nd := msg.Data().(nom.NodeDisconnected)
	return bh.MappedCells{{nodeDriversDict, bh.Key(nd.Node.ID)}}
}

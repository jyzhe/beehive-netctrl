package main

import (
    "flag"
    "log"
    "os"
    "runtime/pprof"

    bh "github.com/kandoo/beehive"
    "github.com/kandoo/beehive-netctrl/controller"
    "github.com/jyzhe/beehive-netctrl/discovery"
    "github.com/kandoo/beehive-netctrl/openflow"
    "github.com/jyzhe/beehive-netctrl/routing"
    // "github.com/jyzhe/beehive-netctrl/switching"
)

var controllertype = flag.String("controller", "master", "specify the controller (master, area1, area2)")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
    flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

    h := bh.NewHive()
    openflow.StartOpenFlow(h)
    controller.RegisterNOMController(h)
    discovery.RegisterDiscovery(h)

    // Register a switch:
    // switching.RegisterSwitch(h, bh.Persistent(1))
    // or a hub:
    // switching.RegisterHub(h, bh.NonTransactional())
	if *controllertype == "master" {
		routing.InstallMasterC(h, bh.Persistent(1))
	} else if *controllertype == "area1" {
		routing.InstallR1(h, bh.Persistent(1))
	} else if *controllertype == "area2"{
		routing.InstallR2(h, bh.Persistent(1))
    } else if *controllertype == "area3"{
        routing.InstallR3(h, bh.Persistent(1))
    } else if *controllertype == "area4"{
        routing.InstallR4(h, bh.Persistent(1))
    }

    // routing.InstallLoadBalancer(h, bh.Persistent(1))
    // routing.InstallRouterIP(h,  bh.Persistent(1))
    h.Start()
}

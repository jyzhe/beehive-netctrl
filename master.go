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

	routing.InstallMaster (h, bh.Persistent(1))
	// routing.InstallLoadBalancer(h, bh.Persistent(1))
	h.Start()
}

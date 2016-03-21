package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"time"

	bh "github.com/kandoo/beehive"
	"github.com/kandoo/beehive-netctrl/controller"
	"github.com/jyzhe/beehive-netctrl/discovery"
	"github.com/kandoo/beehive-netctrl/openflow"
	"github.com/jyzhe/beehive-netctrl/routing"
	// "github.com/jyzhe/beehive-netctrl/switching"
)

var from = flag.Int("from", 1, "First pod in [1..k]")
var to = flag.Int("to", 4, "Last pod in [1..k]")
var k = flag.Int("k", 4, "Number of ports of switches (must be even)")
var epoc = flag.Duration("epoc", 100*time.Millisecond,
	"The duration between route advertisement epocs.")
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

	routing.InstallRouter(h, bh.Persistent(1))
	// routing.InstallLoadBalancer(h, bh.Persistent(1))
	h.Start()
}

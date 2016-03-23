#!/usr/bin/python

"""
BeeHive - Load Balancing and Routing
Author: Jay Yu
February 5th, 2016

Start up a fat tree topology with multiple paths among hosts. The tree has a
3-layer hierarchical structure.
"""

import sys
from time import sleep, time

from mininet.net import Mininet
from mininet.topo import Topo
from mininet.log import setLogLevel
from mininet.node import OVSSwitch, Controller, RemoteController
from mininet.cli import CLI
from mininet.util import pmonitor
from signal import SIGINT

class ThreeLayerTopo(Topo):

    def __init__(self, k=2, **opts):

        Topo.__init__(self, **opts)

        self.k = k
        switchCount = 1

        hosts   = []
        aggr    = []
        tor     = []
        core    = []
        # Create the pods
        for i in range(k):

            core.append(self.addSwitch('s%s' % switchCount))
            switchCount += 1

            # Adding aggregation switches
            for j in range(2):
                aggr.append(self.addSwitch('s%d'%(switchCount)))
                switchCount += 1

            # Adding ToR switches
            for j in range(2):
                tor.append(self.addSwitch('s%d'%(switchCount)))
                switchCount += 1

            # Adding hosts
            for j in range(4):
                hosts.append(self.addHost('h%d'%(i * 4 + j)))

        # Adding the links
        for i in range(k):

            # Connect hosts to tor
            for j in range(4):
                self.addLink(hosts[i * 4 + j], tor[i * 2 + j // 2])

            # Connect tor to aggr
            self.addLink(aggr[i * 2], tor[i * 2])
            self.addLink(aggr[i * 2 + 1], tor[i * 2])
            self.addLink(aggr[i * 2], tor[i * 2 + 1])
            self.addLink(aggr[i * 2 + 1], tor[i * 2 + 1])

        # Connect pods to core
        for i in range(k * 2):

            idx = 0 if i % 2 == 0 else k / 2
            for j in range(k / 2):
                self.addLink(core[j + idx], aggr[i])

topos = { 'mytopo': ( lambda: ThreeLayerTopo() ) }

class MultiSwitch(OVSSwitch):

    map = {}

    def init_map(self, controllers):
        for i in range(2):
            for j in range(1, 6):
                self.map["s%d"%(i * 5 + j)] = controllers[i]


    def start(self, controllers):
        self.init_map(controllers)
        return OVSSwitch.start(self, [self.map[self.name]])

def start_mininet():

    fattree = ThreeLayerTopo()

    c0 = RemoteController(name="beehive-netctrl", ip="127.0.0.1", port=6633)
    c1 = RemoteController(name="beehive-netctrl2", ip="127.0.0.1", port=6634)
    c2 = RemoteController(name="beehive-master", ip="127.0.0.1", port=9080)

    net = Mininet(topo=fattree, switch=MultiSwitch, build=False,
                  autoSetMacs=True, autoStaticArp=True)


    for c in [c0, c1]:
        net.addController(c)

    net.addController(c2)

    net.build()
    net.start()

    # print "Performing prelimilary testing...."
    # pre_test(net)

    CLI(net)
    net.stop()

def pre_test(net):

    print "Waiting for the controller to complete handshake..."
    sleep(5)

    popens = {}
    hosts = net.hosts

    print "Flow 1: h0 starting to ping h6..."
    popens[hosts[0]] = hosts[0].popen('ping', hosts[6].IP())
    sleep(1)
    print "Flow 2: h1 starting to ping h7..."
    popens[hosts[1]] = hosts[1].popen('ping', hosts[7].IP())

    endTime = time() + 10
    for h, line in pmonitor( popens, timeoutms=500 ):
        if h:
           print 'Flow %d: %s' % ( hosts.index(h) + 1, line ),
        if time() >= endTime:
           for p in popens.values():
             p.send_signal( SIGINT )

if __name__ == "__main__":

    setLogLevel('info')

    start_mininet()

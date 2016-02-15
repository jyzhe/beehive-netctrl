#!/usr/bin/python

"""
BeeHive - Load Balancing and Routing
Author: Jay Yu
February 5th, 2016

Start up a fat tree topology with multiple paths among hosts. The tree has a
3-layer hierarchical structure.
"""

from mininet.topo import Topo

class ThreeLayerTopo(Topo):

    def __init__(self, k=4, **opts):

        Topo.__init__(self, **opts)

        self.k = k

        hosts   = []
        aggr    = []
        tor     = []
        core    = []

        for i in range(k):
            core.append(self.addSwitch('c%s' % i))

        for i in range(k):

            # Adding aggregation switches
            for j in range(2):
                aggr.append(self.addSwitch('a%d'%(i * 2 + j)))

            # Adding ToR switches
            for j in range(2):
                tor.append(self.addSwitch('t%d'%(i * 2 + j)))

            # Adding hosts
            for j in range(4):
                hosts.append(self.addHost('h%d'%(i * 4 + j)))

            # Connect hosts to tor
            for j in range(4):
                self.addLink(hosts[i * 4 + j], tor[i * 2 + j // 2])

            # Connect tor to aggr
            self.addLink(aggr[i * 2], tor[i * 2])
            self.addLink(aggr[i * 2 + 1], tor[i * 2])
            self.addLink(aggr[i * 2], tor[i * 2 + 1])
            self.addLink(aggr[i * 2 + 1], tor[i * 2 + 1])

        # Connect aggr to core
        for i in range(k * 2):

            idx = 0 if i % 2 == 0 else k / 2
            for j in range(k / 2):
                self.addLink(core[j + idx], aggr[i])

topos = { 'mytopo': ( lambda: ThreeLayerTopo() ) }

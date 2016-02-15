#!/usr/bin/python

"""
BeeHive - Load Balancing and Routing
Author: Jay Yu
February 5th, 2016

Start up a fat tree topology with multiple paths among hosts. The tree has a
3-layer hierarchical structure.
"""

from mininet.topo import Topo

class SimpleTopo(Topo):

    def __init__(self, k=4, **opts):

        Topo.__init__(self, **opts)

        self.k = k

        hosts = []
        sw = []

        for i in range(4):
            hosts.append(self.addHost('h%d'%i))

        for i in range(4):
            sw.append(self.addSwitch('s%d'%i))

        for i in range(4):
            self.addLink(hosts[i], sw[i])

        for i in range(4):
            for j in range(4):
                if i != j:
                    self.addLink(sw[i], sw[j])

topos = { 'simple': ( lambda: SimpleTopo() ) }

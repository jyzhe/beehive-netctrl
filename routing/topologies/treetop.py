from mininet.topo import Topo

class TreeTopTopo( Topo ):
    "Topology for a tree network with a given depth and fanout."

    def __init__( self, depth=1, fanout=2 ):
        super( TreeTopTopo, self ).__init__()
        # Numbering:  h1..N, s1..M
        self.hostNum = 1
        self.switchNum = 1
        # Build topology
        self.addTree( depth, fanout )

    def addTree( self, depth, fanout ):
        """Add a subtree starting with node n.
           returns: last node added"""
        isSwitch = depth > 0
        if isSwitch:
            node = self.addSwitch( 's%s' % self.switchNum )
            if self.switchNum == 1:
                c = self.addHost( 'h0' )
                self.addLink( node, c )
            self.switchNum += 1
            for _ in range( fanout ):
                child = self.addTree( depth - 1, fanout )
                self.addLink( node, child )
        else:
            node = self.addHost( 'h%s' % self.hostNum )
            self.hostNum += 1
        return node

topos = { 'treetop': TreeTopTopo }

#!/bin/bash

sudo mn --custom $GOPATH/src/github.com/jyzhe/beehive-netctrl/routing/topologies/$1.py --topo $1 --controller remote --arp --mac --switch ovsk,protocols=OpenFlow12

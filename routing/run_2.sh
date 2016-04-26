#!/bin/bash
cd ~/go/workspace/src/github.com/jyzhe/beehive-netctrl/
rm -rf /tmp/beehive*
go run main-kenan.go -of.addr 0.0.0.0:9088 &
go run main2-kenan.go -addr localhost:7678 -statepath /tmp/beehive2 -paddrs localhost:7677 &
go run main3-kenan.go -addr localhost:7679 -of.addr 0.0.0.0:6634 -statepath /tmp/beehive3 -paddrs localhost:7677 &
python routing/topologies/fattree.py

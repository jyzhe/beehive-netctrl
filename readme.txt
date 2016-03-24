To clear the previous instances of beehive and their states
    rm -rf /tmp/beehive*

To run master controller:
    go run master.go -of.addr 0.0.0.0:9080

In a separate window, run the first controller for area 1:
    go run router.go -addr localhost:7678 -statepath /tmp/beehive2 -paddrs localhost:7677

In a separate window, runt he second controller for area 2:
    go run loadbalancer.go -addr localhost:7679 -of.addr 0.0.0.0:6634 -statepath /tmp/beehive3 -paddrs localhost:7677

Then start Mininet (in the topologies directory):
    sudo python fattree.py

You may also run controllers in background using &, for example;
    go run master.go -of.addr 0.0.0.0:9080 &

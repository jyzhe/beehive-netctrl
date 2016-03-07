package nom

import (
	"encoding/gob"
	"fmt"
)

// Special ports.
const (
	PortFlood UID = "Ports.PortBcast"
	PortAll   UID = "Ports.PortAll"
)

// PacketIn messages are emitted when a packet is forwarded to the controller.
type PacketIn struct {
	Node     UID
	InPort   UID
	BufferID PacketBufferID
	Packet   Packet
}

func (in PacketIn) String() string {
	return fmt.Sprintf("packet in on switch %s port %s", in.Node, in.InPort)
}

// PacketOut messages are emitted to send a packet out of a port.
type PacketOut struct {
	Node     UID
	InPort   UID
	BufferID PacketBufferID
	Packet   Packet
	Actions  []Action
}

// Packet is simply the packet data.
type Packet []byte

// DstMAC returns the destination MAC address from the ethernet header.
func (p Packet) DstMAC() MACAddr {
	return MACAddr{p[0], p[1], p[2], p[3], p[4], p[5]}
}

// SrcMAC returns the source MAC address from the ethernet header.
func (p Packet) SrcMAC() MACAddr {
	return MACAddr{p[6], p[7], p[8], p[9], p[10], p[11]}
}

func (p Packet) SrcIP() IPv4Addr {
	return IPv4Addr{p[26], p[27], p[28], p[29]}
}

func (p Packet) DstIP() IPv4Addr {
	return IPv4Addr{p[30], p[31], p[32], p[33]}
}
// PacketBufferID represents a packet buffered in the switch.
type PacketBufferID uint32

func init() {
	gob.Register(Packet{})
	gob.Register(PacketBufferID(0))
	gob.Register(PacketIn{})
	gob.Register(PacketOut{})
}

// Package rtpdump implements the RTPDump file format documented at
// https://www.cs.columbia.edu/irt/software/rtptools/
package rtpdump

import (
	"net"
	"time"

	"github.com/pions/rtcp"
	"github.com/pions/rtp"
)

type Header struct {
	Start  time.Time
	Source net.IP
	Port   int
}

type header struct {
	// start of recording (GMT)
	StartSec  uint32
	StartUsec uint32
	// network source (multicast address)
	Source uint32
	// UDP port
	Port uint16
	// 2 bytes of padding
	_ uint16
}

func (h header) Header() Header {
	start := time.Unix(int64(h.StartSec), int64(h.StartUsec)*1e3)
	source := net.IPv4(
		byte(h.Source>>24),
		byte(h.Source>>16),
		byte(h.Source>>8),
		byte(h.Source),
	)
	return Header{
		Start:  start,
		Source: source,
		Port:   int(h.Port),
	}
}

type Packet struct {
	// milliseconds since the start of recording
	Offset uint32

	IsRTCP bool
	RTCP   rtcp.Packet
	RTP    *rtp.Packet
}

type packetHeader struct {
	// length of packet, including this header (may be smaller than
	// plen if not whole packet recorded)
	Length uint16
	// Actual header+payload length for RTP, 0 for RTCP
	PacketLength uint16
	// milliseconds since the start of recording
	Offset uint32
}

const pktHeaderLen = 8

func timestamp(t time.Time) (sec, usec uint32) {
	ns := t.UnixNano()
	sec = uint32(ns / int64(time.Second))
	usec = uint32((ns % int64(time.Second)) / int64(time.Microsecond))
	return sec, usec
}

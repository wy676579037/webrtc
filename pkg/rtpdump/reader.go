package rtpdump

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/pions/rtcp"
	"github.com/pions/rtp"
)

type Reader struct {
	reader io.Reader

	header Header
}

func (r *Reader) Header() Header {
	return r.header
}

func NewReader(r io.Reader) (*Reader, error) {
	bio := bufio.NewReader(r)

	preamble, err := bio.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(preamble, "#!rtpplay1.0 ") {
		return nil, errors.New("invalid header")
	}

	var h header
	if err := binary.Read(bio, binary.BigEndian, &h); err != nil {
		return nil, fmt.Errorf("failed to read header: %v", err)
	}

	return &Reader{
		reader: bio,
		header: h.Header(),
	}, nil
}

func (r *Reader) Next() (*Packet, error) {
	var h packetHeader
	if err := binary.Read(r.reader, binary.BigEndian, &h); err != nil {
		return nil, err
	}

	pkt := &Packet{
		Offset: h.Offset,
	}

	buf := make([]byte, h.Length-pktHeaderLen)

	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return nil, err
	}

	if h.PacketLength > 0 {
		pkt.IsRTCP = false

		var rtpPkt rtp.Packet
		if err := rtpPkt.Unmarshal(buf); err != nil {
			return nil, err
		}
		pkt.RTP = &rtpPkt

	} else {
		pkt.IsRTCP = true

		rtcpPkt, err := rtcp.Unmarshal(buf)
		if err != nil {
			return nil, err
		}
		pkt.RTCP = rtcpPkt
	}

	return pkt, nil
}

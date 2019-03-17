package rtpdump

import (
	"encoding/binary"
	"io"
	"sync"
	"time"

	"github.com/pions/rtcp"
	"github.com/pions/rtp"
)

var preamble = []byte("#!rtpplay1.0 0.0.0.0/0\n")

type Writer struct {
	writer io.Writer

	start   time.Time
	startMu sync.Mutex
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) headerWritten() bool {
	w.startMu.Lock()
	defer w.startMu.Unlock()

	return !w.start.IsZero()
}

func (w *Writer) writeHeader() error {
	w.startMu.Lock()
	defer w.startMu.Unlock()

	w.start = time.Now()

	sec, usec := timestamp(time.Now())
	hdr := header{
		StartSec:  sec,
		StartUsec: usec,
		Source:    0,
		Port:      0,
	}

	if _, err := w.writer.Write(preamble); err != nil {
		return err
	}

	return binary.Write(w.writer, binary.BigEndian, hdr)
}

func (w *Writer) WriteRTP(p *rtp.Packet) error {
	payload, err := p.Marshal()
	if err != nil {
		return err
	}

	return w.writePacket(payload, false)
}

func (w *Writer) WriteRTCP(p rtcp.Packet) error {
	payload, err := p.Marshal()
	if err != nil {
		return err
	}

	return w.writePacket(payload, true)
}

func (w *Writer) writePacket(payload []byte, rtcp bool) error {
	if !w.headerWritten() {
		if err := w.writeHeader(); err != nil {
			return err
		}
	}

	w.startMu.Lock()
	defer w.startMu.Unlock()
	offset := time.Since(w.start)
	offsetMs := offset.Nanoseconds() / int64(time.Millisecond)

	pLen := uint16(len(payload))
	if rtcp {
		pLen = 0
	}

	hdr := packetHeader{
		Length:       uint16(len(payload) + pktHeaderLen),
		PacketLength: pLen,
		Offset:       uint32(offsetMs),
	}
	if err := binary.Write(w.writer, binary.BigEndian, hdr); err != nil {
		return err
	}

	if _, err := w.writer.Write(payload); err != nil {
		return err
	}

	return nil
}

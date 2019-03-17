package rtpdump

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pions/webrtc"
)

type Player struct {
	reader *Reader

	track *webrtc.Track

	done chan struct{}

	start   time.Time
	startMu sync.Mutex
}

func NewPlayer(r io.Reader, track *webrtc.Track) (*Player, error) {
	reader, err := NewReader(r)
	if err != nil {
		return nil, err
	}
	return &Player{
		reader: reader,
		done:   make(chan struct{}),
		track:  track,
	}, nil
}

func (p *Player) loop() {
	for {
		select {
		case <-p.done:
			return
		default:
		}

		pkt, err := p.reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
		}

		currentTime := time.Since(p.start)
		pts := time.Duration(pkt.Offset) * time.Millisecond

		if currentTime < pts {
			select {
			case <-time.After(pts - currentTime):
				// ready
			case <-p.done:
				return
			}
		}

		if pkt.IsRTCP {
			// ignore for now
		} else {
			if err := p.track.WriteRTP(pkt.RTP); err != nil {
				fmt.Println("[error]", err)
			}
		}
	}
}

func (p *Player) Start() {
	p.startMu.Lock()
	defer p.startMu.Unlock()

	if !p.start.IsZero() {
		return // already playing
	}

	p.start = time.Now()
	p.loop()
}

func (p *Player) Stop() {
	p.startMu.Lock()
	defer p.startMu.Unlock()

	p.start = time.Time{}
	p.done <- struct{}{}
	// TODO: block until stopped
}

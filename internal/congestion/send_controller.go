package congestion

import (
	"time"
)

// SendController integrates BBRv2, pacer, and rate sampler
type SendController struct {
	sampler          *Sampler
	bbr              *BBRv2
	pacer            *Pacer
	congestionWindow int
	mtu              int
}

// NewSendController creates a new send controller
func NewSendController(mtu int, initialCWND int) *SendController {
	return &SendController{
		sampler:          NewSampler(),
		bbr:              NewBBRv2(mtu, initialCWND),
		pacer:            NewPacer(mtu),
		congestionWindow: initialCWND,
		mtu:              mtu,
	}
}

// OnPacketSent is called when a packet is sent
func (sc *SendController) OnPacketSent(now time.Time, size int, isAppLimited bool) {
	sc.sampler.OnPacketSent(now, size, isAppLimited)
}

// OnAck is called when an ACK is received
func (sc *SendController) OnAck(now time.Time, ackedBytes int, rtt time.Duration) {
	rs := sc.sampler.OnAck(now, ackedBytes)
	cwnd, pace := sc.bbr.OnAck(Sample{RS: rs, RTT: rtt})
	sc.congestionWindow = cwnd
	sc.pacer.SetRate(pace)
}

// OnLoss is called when packet loss is detected
func (sc *SendController) OnLoss(bytesLost int) {
	cwnd, pace := sc.bbr.OnLoss()
	sc.congestionWindow = cwnd
	sc.pacer.SetRate(pace)
}

// CanSend checks if a packet can be sent (pacing + congestion window)
func (sc *SendController) CanSend(now time.Time, size int) bool {
	// Check pacing
	if !sc.pacer.Allow(now, size) {
		return false
	}

	// Check congestion window
	return sc.congestionWindow >= size
}

// GetCWND returns the current congestion window
func (sc *SendController) GetCWND() int {
	return sc.congestionWindow
}

// GetPacingRate returns the current pacing rate
func (sc *SendController) GetPacingRate() int64 {
	return sc.pacer.GetRate()
}

// GetBandwidth returns the current bandwidth estimate
func (sc *SendController) GetBandwidth() float64 {
	return sc.bbr.GetBandwidth()
}

// GetMinRTT returns the minimum RTT
func (sc *SendController) GetMinRTT() time.Duration {
	return sc.bbr.GetMinRTT()
}

// GetState returns the current BBRv2 state
func (sc *SendController) GetState() bbrState {
	return sc.bbr.GetState()
}

package congestion

import (
	"time"
)

// Pacer implements token bucket pacing for BBRv2
type Pacer struct {
	rateBps  int64
	tokens   float64
	lastTick time.Time
	mtu      int
}

// NewPacer creates a new pacer
func NewPacer(mtu int) *Pacer {
	return &Pacer{mtu: mtu}
}

// SetRate sets the pacing rate in bytes per second
func (p *Pacer) SetRate(bps int64) {
	if bps < 0 {
		bps = 0
	}
	p.rateBps = bps
}

// Allow checks if a packet of given size can be sent now
func (p *Pacer) Allow(now time.Time, size int) bool {
	if p.lastTick.IsZero() {
		p.lastTick = now
	}

	elapsed := now.Sub(p.lastTick).Seconds()
	p.lastTick = now

	// Add tokens based on elapsed time
	p.tokens += float64(p.rateBps) * elapsed

	// Limit burst to 10 MTUs
	maxBurst := float64(10 * p.mtu)
	if p.tokens > maxBurst {
		p.tokens = maxBurst
	}

	need := float64(size)
	if p.tokens >= need {
		p.tokens -= need
		return true
	}

	return false
}

// GetRate returns the current pacing rate
func (p *Pacer) GetRate() int64 {
	return p.rateBps
}

// GetTokens returns the current token count
func (p *Pacer) GetTokens() float64 {
	return p.tokens
}


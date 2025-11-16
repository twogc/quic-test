package congestion

import (
	"fmt"
	"time"
	
	"go.uber.org/zap"
)

var debugLogger *zap.Logger

func init() {
	var err error
	debugLogger, err = zap.NewDevelopment()
	if err != nil {
		// Fallback to no-op logger
		debugLogger = zap.NewNop()
	}
}

// SetDebugLogger sets the debug logger for congestion control
func SetDebugLogger(logger *zap.Logger) {
	debugLogger = logger
}

// CongestionController interface for different CC algorithms
type CongestionController interface {
	OnAck(Sample) (cwnd int, pacing int64)
	OnLoss() (cwnd int, pacing int64)
	GetCWND() int
	GetPacingRate() int64
	GetBandwidth() float64
	GetMinRTT() time.Duration
	SetQlogCallback(func(eventType string, data map[string]interface{}))
}

// BBRv2Controller wraps BBRv2 to implement CongestionController
type BBRv2Controller struct {
	*BBRv2
}

// BBRv3Controller wraps BBRv3 to implement CongestionController
type BBRv3Controller struct {
	*BBRv3
}

func (c *BBRv3Controller) OnAck(s Sample) (cwnd int, pacing int64) {
	c.BBRv3.OnPacketAcked()
	return c.BBRv3.OnAck(s)
}

func (c *BBRv3Controller) OnLoss() (cwnd int, pacing int64) {
	return c.BBRv3.OnLoss()
}

// SendController integrates congestion control, pacer, and rate sampler
type SendController struct {
	sampler          *Sampler
	cc               CongestionController
	pacer            *Pacer
	congestionWindow int
	mtu              int
	algorithm        string // "bbrv2", "bbrv3", etc.
}

// NewSendController creates a new send controller with specified algorithm
func NewSendController(mtu int, initialCWND int, algorithm string) *SendController {
	sc := &SendController{
		sampler:          NewSampler(),
		pacer:            NewPacer(mtu),
		congestionWindow: initialCWND,
		mtu:              mtu,
		algorithm:        algorithm,
	}
	
	// Initialize congestion controller based on algorithm
	switch algorithm {
	case "bbrv3":
		bbrv3 := NewBBRv3(mtu, initialCWND)
		sc.cc = &BBRv3Controller{bbrv3}
	case "bbrv2", "bbr":
		bbrv2 := NewBBRv2(mtu, initialCWND)
		sc.cc = &BBRv2Controller{bbrv2}
	default:
		// Default to BBRv2
		bbrv2 := NewBBRv2(mtu, initialCWND)
		sc.cc = &BBRv2Controller{bbrv2}
		sc.algorithm = "bbrv2"
	}
	
	return sc
}

// OnPacketSent is called when a packet is sent
func (sc *SendController) OnPacketSent(now time.Time, size int, isAppLimited bool) {
	defer func() {
		if r := recover(); r != nil {
			debugLogger.Error("Panic in OnPacketSent",
				zap.String("error", fmt.Sprintf("%v", r)),
				zap.Int("size", size))
			panic(r)
		}
	}()
	
	sc.sampler.OnPacketSent(now, size, isAppLimited)
	
	// Track packet sending for BBRv3
	if bbrv3, ok := sc.cc.(*BBRv3Controller); ok {
		bbrv3.OnPacketSent()
		// Логируем только при ошибках
		// debugLogger.Debug("SendController.OnPacketSent (BBRv3)",
		// 	zap.Int("size", size),
		// 	zap.Bool("isAppLimited", isAppLimited))
	}
}

// OnAck is called when an ACK is received
func (sc *SendController) OnAck(now time.Time, ackedBytes int, rtt time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			debugLogger.Error("Panic in OnAck",
				zap.String("error", fmt.Sprintf("%v", r)),
				zap.Int("ackedBytes", ackedBytes),
				zap.Duration("rtt", rtt))
			panic(r) // Re-panic after logging
		}
	}()
	
	rs := sc.sampler.OnAck(now, ackedBytes)
	
	// Логируем только ошибки или раз в 100 ACK
	// debugLogger.Debug("SendController.OnAck",
	// 	zap.Int("ackedBytes", ackedBytes),
	// 	zap.Duration("rtt", rtt),
	// 	zap.Float64("bandwidth_bps", rs.BandwidthBps()),
	// 	zap.String("algorithm", sc.algorithm))
	
	cwnd, pace := sc.cc.OnAck(Sample{RS: rs, RTT: rtt})
	
	// Проверяем на некорректные значения
	if cwnd <= 0 {
		debugLogger.Warn("SendController.OnAck: invalid cwnd",
			zap.Int("cwnd", cwnd),
			zap.Int("ackedBytes", ackedBytes))
		cwnd = sc.mtu * 10 // Safe default
	}
	if pace <= 0 {
		debugLogger.Warn("SendController.OnAck: invalid pacing rate",
			zap.Int64("pace", pace),
			zap.Int("ackedBytes", ackedBytes))
		pace = 1000000 // 1 Mbps safe default
	}
	
	sc.congestionWindow = cwnd
	sc.pacer.SetRate(pace)
}

// OnLoss is called when packet loss is detected
func (sc *SendController) OnLoss(bytesLost int) {
	cwnd, pace := sc.cc.OnLoss()
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
	return sc.cc.GetBandwidth()
}

// GetMinRTT returns the minimum RTT
func (sc *SendController) GetMinRTT() time.Duration {
	return sc.cc.GetMinRTT()
}

// GetState returns the current BBR state (works for both v2 and v3)
func (sc *SendController) GetState() string {
	switch sc.algorithm {
	case "bbrv3":
		if bbrv3, ok := sc.cc.(*BBRv3Controller); ok {
			return bbrv3.getStateString()
		}
	case "bbrv2", "bbr":
		if bbrv2, ok := sc.cc.(*BBRv2Controller); ok {
			return bbrv2.getStateString()
		}
	}
	return "Unknown"
}

// GetAlgorithm returns the current congestion control algorithm
func (sc *SendController) GetAlgorithm() string {
	return sc.algorithm
}

// GetBBRv3Metrics returns BBRv3-specific metrics if using BBRv3
func (sc *SendController) GetBBRv3Metrics() (*BBRv3Metrics, bool) {
	if bbrv3, ok := sc.cc.(*BBRv3Controller); ok {
		metrics := bbrv3.GetMetrics()
		return &metrics, true
	}
	return nil, false
}

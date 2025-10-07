package congestion

import (
	"time"
)

// RateSample представляет образец пропускной способности
type RateSample struct {
	Delivered    int64         // суммарно подтверждено байт к моменту отправки пакета
	DeliveredAt  time.Time     // время, когда measured delivered зафиксирован
	FirstSentAt  time.Time     // время отправки пакета, который «отмеряет» интервал
	Interval     time.Duration // DeliveredAt - FirstSentAt (с защитами)
	BytesAcked   int64         // количество байт в этом ACK
	IsAppLimited bool          // был ли пакет отправлен в состоянии app-limited
}

// Sampler измеряет delivery rate для BBRv2
type Sampler struct {
	delivered   int64     // общее количество доставленных байт
	deliveredAt time.Time // время последнего обновления delivered
	firstSentAt time.Time // время отправки первого пакета в интервале
	appLimited  bool      // флаг app-limited состояния
}

// NewSampler создает новый sampler
func NewSampler() *Sampler {
	return &Sampler{}
}

// OnPacketSent вызывается при отправке пакета
func (s *Sampler) OnPacketSent(now time.Time, size int, isAppLimited bool) {
	if s.firstSentAt.IsZero() {
		s.firstSentAt = now
	}

	// Обновляем флаг app-limited
	if isAppLimited {
		s.appLimited = true
	}
}

// OnAck вызывается при получении ACK
func (s *Sampler) OnAck(now time.Time, ackedBytes int) RateSample {
	s.delivered += int64(ackedBytes)

	rs := RateSample{
		Delivered:    s.delivered,
		DeliveredAt:  now,
		FirstSentAt:  s.firstSentAt,
		Interval:     now.Sub(s.firstSentAt),
		BytesAcked:   int64(ackedBytes),
		IsAppLimited: s.appLimited,
	}

	// Защита от нулевых/слишком коротких интервалов
	if rs.Interval < 1*time.Millisecond {
		rs.Interval = 1 * time.Millisecond
	}

	// При первом ACK сдвигаем начало интервала
	s.firstSentAt = now
	s.appLimited = false

	return rs
}

// BandwidthBps возвращает пропускную способность в байтах в секунду
func (rs *RateSample) BandwidthBps() float64 {
	if rs.Interval <= 0 {
		return 0
	}
	return float64(rs.BytesAcked) / rs.Interval.Seconds()
}

// BandwidthMbps возвращает пропускную способность в мегабитах в секунду
func (rs *RateSample) BandwidthMbps() float64 {
	return rs.BandwidthBps() * 8 / (1024 * 1024)
}

// IsValid проверяет, валиден ли образец
func (rs *RateSample) IsValid() bool {
	return rs.Interval > 0 && rs.BytesAcked > 0
}

// Reset сбрасывает состояние sampler
func (s *Sampler) Reset() {
	s.delivered = 0
	s.deliveredAt = time.Time{}
	s.firstSentAt = time.Time{}
	s.appLimited = false
}

// GetDelivered возвращает текущее количество доставленных байт
func (s *Sampler) GetDelivered() int64 {
	return s.delivered
}

// IsAppLimited возвращает, находимся ли мы в app-limited состоянии
func (s *Sampler) IsAppLimited() bool {
	return s.appLimited
}

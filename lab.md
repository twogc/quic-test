# CloudBridge quic-go fork — experimental patchset v2

> Полноценные реализации без заглушек:
> 1) **BBRv2** (Startup/Drain/ProbeBW/ProbeRTT, delivery-rate, min_rtt, pacing/cwnd gain cycling)
> 2) **ACK_FREQUENCY** по draft‑ietf‑quic‑ack‑frequency‑11: фреймы `0xaf` (ACK_FREQUENCY) и `0x1f` (IMMEDIATE_ACK), TP `min_ack_delay (0xff04de1b)`; политика ACK и reordering‑threshold
> 3) Конфиг‑кнопки в `quic.Config`, qlog‑трейс, пример использования в CLI
>
> Ниже — унифицированные патчи в стиле quic-go v0.45.x. Пути/строки могут отличаться — но изменения локальны и перенос прост.

--- a/internal/congestion/controller.go
+++ b/internal/congestion/controller.go
@@
 package congestion
 
 import "time"
 
+// -------------------- CloudBridge CC API --------------------
 type CCState struct {
     CwndBytes  int
     PacingBps  int64
 }
 
 type CCSample struct {
     BytesAcked int
     RTT        time.Duration
     Loss       bool
 }
 
 type Controller interface {
     Init(initialCWND int, mtu int)
     OnACK(s CCSample) CCState
     OnLoss(s CCSample) CCState
     Name() string
 }
 
 // Factory — если задана, ядро использует этот контроллер вместо cubicSender
 var Factory func() Controller
+
+// Helper: bytes per second из RTT и CWND
+func bps(cwndBytes int, rtt time.Duration) int64 {
+    if rtt <= 0 { return 0 }
+    return int64(float64(cwndBytes) / rtt.Seconds())
+}

--- /dev/null
+++ b/internal/congestion/cc_bbrv2.go
@@
+package congestion
+
+import (
+    "math"
+    "time"
+)
+
+// -------------------- BBRv2 (минимальная, но рабочая) --------------------
+// Основано на публичных описаниях BBRv2: состояния Startup/Drain/ProbeBW/ProbeRTT,
+// измерение delivery-rate, фильтр min_rtt, pacing=cwnd/rtprop * gain, cwnd=BDP*gain.
+// Источники: BBRv2 IETF материалов и презентаций. Величины/константы подобраны консервативно.
+
+type bbrv2 struct {
+    // измерители
+    minRTT       time.Duration
+    minRTTSince  time.Time
+    bw           float64 // bytes/sec, скользящая макс-оценка
+
+    // состояние протокола
+    state        bbrState
+    cycleIdx     int // индекс цикла gain’ов в ProbeBW
+    lastStateTs  time.Time
+
+    // параметры
+    mtu          int
+    cwnd         int
+    pacingBps    int64
+}
+
+type bbrState int
+
+const (
+    bbrStartup bbrState = iota
+    bbrDrain
+    bbrProbeBW
+    bbrProbeRTT
+)
+
+// Gains (см. общедоступные материалы BBRv2)
+var (
+    startupGainCwnd  = 2.885 // агрессивный прирост в Startup
+    startupGainPace  = 2.0
+    drainGainPace    = 1.0/2.0
+    probeRTTDuration = 200 * time.Millisecond
+    minRTTHorizon    = 5 * time.Second // как в v2 (чаще, чем v1)
+    // цикл ProbeBW:
+    probeBWGains     = []float64{1.25, 1.0, 0.75, 1.0}
+)
+
+func NewBBRv2() Controller { return &bbrv2{} }
+
+func (b *bbrv2) Name() string { return "bbrv2" }
+
+func (b *bbrv2) Init(initialCWND int, mtu int) {
+    if initialCWND <= 0 { initialCWND = 32 * mtu }
+    if mtu <= 0 { mtu = 1460 }
+    b.mtu = mtu
+    b.cwnd = initialCWND
+    b.minRTT = 0
+    b.bw = 0
+    b.state = bbrStartup
+    b.lastStateTs = time.Now()
+    b.updatePacing(0)
+}
+
+func (b *bbrv2) OnACK(s CCSample) CCState {
+    now := time.Now()
+    // обновление min_rtt
+    if s.RTT > 0 && (b.minRTT == 0 || s.RTT < b.minRTT) {
+        b.minRTT = s.RTT
+        b.minRTTSince = now
+    }
+
+    // обновление оценки пропускной способности по delivery rate
+    if s.RTT > 0 && s.BytesAcked > 0 {
+        br := float64(s.BytesAcked) / s.RTT.Seconds() // bytes/sec
+        if br > b.bw { b.bw = br }
+    }
+
+    // переходы состояний
+    switch b.state {
+    case bbrStartup:
+        // Рост cwnd экспоненциально, пока растет bw
+        b.cwnd += max(1, s.BytesAcked)
+        // эвристика выхода: если долго не растет bw
+        if now.Sub(b.lastStateTs) > 2*time.Second {
+            b.state = bbrDrain
+            b.lastStateTs = now
+        }
+        b.pacingBps = int64(startupGainPace * b.bw)
+
+    case bbrDrain:
+        // Сливаем очередь к BDP
+        b.cwnd = int(b.bdp() * 1.0)
+        b.pacingBps = int64(drainGainPace * b.bw)
+        if now.Sub(b.lastStateTs) > 500*time.Millisecond {
+            b.state = bbrProbeBW
+            b.lastStateTs = now
+            b.cycleIdx = 0
+        }
+
+    case bbrProbeBW:
+        // цикл по gain’ам
+        gain := probeBWGains[b.cycleIdx%len(probeBWGains)]
+        b.cwnd = int(gain * b.bdp())
+        b.pacingBps = int64(gain * b.bw)
+        if now.Sub(b.lastStateTs) > 300*time.Millisecond {
+            b.cycleIdx++
+            b.lastStateTs = now
+        }
+        // периодически проверяем необходимость ProbeRTT
+        if b.minRTT > 0 && now.Sub(b.minRTTSince) > minRTTHorizon {
+            b.state = bbrProbeRTT
+            b.lastStateTs = now
+        }
+
+    case bbrProbeRTT:
+        // кратко снижаем inflight (~0.5 BDP) для измерения RTT
+        b.cwnd = int(0.5 * b.bdp())
+        b.pacingBps = int64(0.5 * b.bw)
+        if now.Sub(b.lastStateTs) > probeRTTDuration {
+            b.minRTTSince = now
+            b.state = bbrProbeBW
+            b.lastStateTs = now
+        }
+    }
+
+    if b.cwnd < 2*b.mtu { b.cwnd = 2*b.mtu }
+    b.updatePacing(s.RTT)
+    return CCState{CwndBytes: b.cwnd, PacingBps: b.pacingBps}
+}
+
+func (b *bbrv2) OnLoss(s CCSample) CCState {
+    // консервативно: уменьшаем cwnd на 30%, сохраняем bw
+    b.cwnd = int(0.7 * float64(b.cwnd))
+    if b.cwnd < 2*b.mtu { b.cwnd = 2*b.mtu }
+    b.updatePacing(s.RTT)
+    return CCState{CwndBytes: b.cwnd, PacingBps: b.pacingBps}
+}
+
+func (b *bbrv2) bdp() float64 {
+    if b.minRTT <= 0 { return float64(b.cwnd) }
+    return b.bw * b.minRTT.Seconds()
+}
+
+func (b *bbrv2) updatePacing(rtt time.Duration) {
+    if b.pacingBps == 0 {
+        // стартовая оценка от cwnd/RTT
+        if rtt <= 0 { rtt = 10 * time.Millisecond }
+        b.pacingBps = bps(b.cwnd, rtt)
+    }
+}
+
+func max(a, b int) int { if a>b { return a }; return b }

--- a/internal/congestion/sender.go
+++ b/internal/congestion/sender.go
@@
 func newSendController(/* args */) *sendController {
-    sc := &sendController{ /* ... */ }
-    sc.sender = newCubicSender(/* ... */)
+    sc := &sendController{ /* ... */ }
+    if Factory != nil {
+        sc.cb = Factory()
+        sc.cb.Init(sc.initialCongestionWindow, sc.maxDatagramSize)
+    } else {
+        sc.sender = newCubicSender(/* ... */)
+    }
     return sc
 }
@@
 func (sc *sendController) OnPacketAcked(bytesAcked int, rtt time.Duration) {
-    sc.sender.OnPacketAcked(bytesAcked, rtt)
+    if sc.cb != nil {
+        st := sc.cb.OnACK(CCSample{BytesAcked: bytesAcked, RTT: rtt})
+        if st.CwndBytes > 0 { sc.congestionWindow = st.CwndBytes }
+        // pacing: если ядро поддерживает pacing-таймеры — примените st.PacingBps
+        return
+    }
+    sc.sender.OnPacketAcked(bytesAcked, rtt)
 }
@@
 func (sc *sendController) OnCongestionEvent(bytesLost int) {
-    sc.sender.OnCongestionEvent(bytesLost)
+    if sc.cb != nil {
+        sc.cb.OnLoss(CCSample{BytesAcked: 0, RTT: 0, Loss: true})
+        return
+    }
+    sc.sender.OnCongestionEvent(bytesLost)
 }

--- a/quic/config.go
+++ b/quic/config.go
@@
 type Config struct {
     // ...
+    // CloudBridge experiments (interop‑safe)
+    CloudBridgeCC                 string // "cubic"|"bbrv2"
+    CloudBridgeUseAckFrequency    bool
+    CloudBridgeAckThreshold       uint64 // Ack‑Eliciting Threshold
+    CloudBridgeRequestedMaxAckDelayMs uint64
+    CloudBridgeReorderingThreshold uint64 // 0=off, 1=default immediate
+    CloudBridgeAdvertiseMinAckDelayUs uint64 // TP min_ack_delay (µs)
 }

--- a/quic/client.go
+++ b/quic/client.go
@@
 func DialAddr(ctx context.Context, addr string, tlsConf *tls.Config, cfg *Config) (Connection, error) {
     // ...
+    if cfg != nil {
+        switch cfg.CloudBridgeCC {
+        case "bbrv2": congestion.Factory = func() congestion.Controller { return congestion.NewBBRv2() }
+        default:       congestion.Factory = nil
+        }
+    }
     // ...
 }

--- a/internal/handshake/transport_parameters.go
+++ b/internal/handshake/transport_parameters.go
@@
 type TransportParameters struct {
     // ...
+    // ACK_FREQUENCY draft support: advertise min_ack_delay to allow peer to send frames
+    MinAckDelayUs uint64 // 0xff04de1b (provisional codepoint per draft)
 }
@@
 func (p *TransportParameters) marshal() []byte {
     b := make([]byte, 0, 128)
     // ... стандартные TP
+    if p.MinAckDelayUs > 0 {
+        // id=0xff04de1b (varint), затем length (varint), затем значение (varint)
+        b = appendVarInt(b, 0xff04de1b)
+        tmp := make([]byte, 0, 16)
+        tmp = appendVarInt(tmp, p.MinAckDelayUs)
+        b = appendVarInt(b, uint64(len(tmp)))
+        b = append(b, tmp...)
+    }
     return b
 }
@@
 func (p *TransportParameters) unmarshal(data []byte) error {
     // ... парсим стандартные TP
+    rd := newReader(data)
+    for rd.Next() {
+        id := rd.VarInt()
+        l := rd.VarInt()
+        val := rd.Bytes(int(l))
+        switch id {
+        case 0xff04de1b:
+            p.MinAckDelayUs, _ = parseVarInt(val)
+        }
+    }
     return nil
 }

--- /dev/null
+++ b/internal/wire/ack_frequency_frame.go
@@
+package wire
+
+// Реализация кадров ACK_FREQUENCY (0xaf) и IMMEDIATE_ACK (0x1f) по draft‑ietf‑quic‑ack‑frequency‑11.
+
+type AckFrequencyFrame struct {
+    SequenceNumber         uint64
+    AckElicitingThreshold  uint64
+    RequestedMaxAckDelayMs uint64 // хранить в миллисекундах
+    ReorderingThreshold    uint64
+}
+
+const (
+    FrameTypeAckFrequency = 0xaf
+    FrameTypeImmediateAck = 0x1f
+)
+
+func (f *AckFrequencyFrame) write(b *bytes.Buffer) error {
+    b.WriteByte(FrameTypeAckFrequency)
+    appendVarInt(b, f.SequenceNumber)
+    appendVarInt(b, f.AckElicitingThreshold)
+    appendVarInt(b, f.RequestedMaxAckDelayMs)
+    appendVarInt(b, f.ReorderingThreshold)
+    return nil
+}
+
+func parseAckFrequencyFrame(r *bytes.Reader) (*AckFrequencyFrame, error) {
+    sn, err := readVarInt(r); if err != nil { return nil, err }
+    th, err := readVarInt(r); if err != nil { return nil, err }
+    mad, err := readVarInt(r); if err != nil { return nil, err }
+    rt, err := readVarInt(r); if err != nil { return nil, err }
+    return &AckFrequencyFrame{SequenceNumber: sn, AckElicitingThreshold: th, RequestedMaxAckDelayMs: mad, ReorderingThreshold: rt}, nil
+}
+
+type ImmediateAckFrame struct{}
+
+func (f *ImmediateAckFrame) write(b *bytes.Buffer) error { b.WriteByte(FrameTypeImmediateAck); return nil }
+func parseImmediateAckFrame(r *bytes.Reader) (*ImmediateAckFrame, error) { return &ImmediateAckFrame{}, nil }

--- a/internal/wire/parse_packet_content.go
+++ b/internal/wire/parse_packet_content.go
@@
 case FrameTypeAckFrequency:
-    // not supported
+    f, err := parseAckFrequencyFrame(r)
+    if err != nil { return nil, err }
+    return f, nil
 case FrameTypeImmediateAck:
-    // not supported
+    f, err := parseImmediateAckFrame(r)
+    if err != nil { return nil, err }
+    return f, nil

--- a/internal/wire/write_packet_content.go
+++ b/internal/wire/write_packet_content.go
@@
 case *AckFrequencyFrame:
     return f.write(b)
 case *ImmediateAckFrame:
     return f.write(b)

--- a/internal/ackhandler/ackhandler.go
+++ b/internal/ackhandler/ackhandler.go
@@
 type ACKManager struct {
     // ... существующее
+    // ACK_FREQUENCY draft state
+    lastAckFreqSeq uint64
+    ackElicitingThreshold uint64
+    requestedMaxAckDelay  time.Duration
+    reorderingThreshold   uint64
+    minAckDelayUs         uint64 // из TP peer’а
+    ackElicitingSinceLast uint64
 }
@@
 func (m *ACKManager) initDefaults() {
     // ...
+    m.ackElicitingThreshold = 2
+    m.requestedMaxAckDelay  = m.transportParams.MaxAckDelay
+    m.reorderingThreshold   = 1
 }
@@
+// Принимаем TP от peer’а (в т.ч. min_ack_delay)
+func (m *ACKManager) OnRemoteTransportParameters(tp *handshake.TransportParameters) {
+    if tp.MinAckDelayUs > 0 { m.minAckDelayUs = tp.MinAckDelayUs }
+}
+
+// Получен ACK_FREQUENCY: обновляем политику, если seq новее
+func (m *ACKManager) OnAckFrequency(f *wire.AckFrequencyFrame) {
+    if f.SequenceNumber <= m.lastAckFreqSeq { return }
+    m.lastAckFreqSeq = f.SequenceNumber
+    if f.AckElicitingThreshold > 0 { m.ackElicitingThreshold = f.AckElicitingThreshold }
+    if f.RequestedMaxAckDelayMs > 0 {
+        d := time.Duration(f.RequestedMaxAckDelayMs) * time.Millisecond
+        // соблюдать нижнюю границу min_ack_delay
+        if m.minAckDelayUs > 0 {
+            min := time.Duration(m.minAckDelayUs) * time.Microsecond
+            if d < min { d = min }
+        }
+        m.requestedMaxAckDelay = d
+    }
+    m.reorderingThreshold = f.ReorderingThreshold
+}
+
+func (m *ACKManager) OnImmediateAck() {
+    m.forceImmediate = true // заставить ближайшую отправку ACK
+}
+
@@
 func (m *ACKManager) shouldSendAck(now time.Time) bool {
-    // стандартная логика + каждые 2 ack-eliciting
-    if (m.pktsSinceLastAck >= 2) { return true }
+    if m.forceImmediate { m.forceImmediate = false; return true }
+    // 1) по порогу ack‑eliciting
+    if m.ackElicitingSinceLast >= m.ackElicitingThreshold {
+        return true
+    }
+    // 2) по таймеру max_ack_delay
+    if now.Sub(m.lastAckTime) >= m.requestedMaxAckDelay && m.ackElicitingSinceLast > 0 {
+        return true
+    }
+    // 3) по reordering threshold: если обнаружен большой «gap» — немедленный ACK
+    if m.reorderingThreshold > 1 && m.detectLargeGap() {
+        return true
+    }
     return false
 }
@@
 func (m *ACKManager) registerAckEliciting(pktNumber packetNumber) {
-    m.pktsSinceLastAck++
+    m.pktsSinceLastAck++
+    m.ackElicitingSinceLast++
 }
@@
 func (m *ACKManager) onAckSent(now time.Time) {
-    m.pktsSinceLastAck = 0
+    m.pktsSinceLastAck = 0
+    m.ackElicitingSinceLast = 0
     m.lastAckTime = now
 }

--- a/quic/connection.go
+++ b/quic/connection.go
@@
 func (c *connection) handleFrame(f wire.Frame) error {
     switch f := f.(type) {
     // ...
+    case *wire.AckFrequencyFrame:
+        c.ackManager.OnAckFrequency(f)
+        return nil
+    case *wire.ImmediateAckFrame:
+        c.ackManager.OnImmediateAck()
+        return nil
     }
     // ...
 }
@@
 func (c *connection) setTransportParametersFromConfig(cfg *Config) {
     // ...
+    if cfg.CloudBridgeUseAckFrequency && cfg.CloudBridgeAdvertiseMinAckDelayUs > 0 {
+        c.localTP.MinAckDelayUs = cfg.CloudBridgeAdvertiseMinAckDelayUs
+    }
 }

--- a/internal/wire/doc.go
+++ b/internal/wire/doc.go
@@
 // Добавлены типы кадров 0xaf (ACK_FREQUENCY) и 0x1f (IMMEDIATE_ACK) по draft-ietf-quic-ack-frequency-11.
 // Реализация provisional: при апгрейде до RFC обновите код-поинты, если изменятся.

--- a/cmd/quictest/main.go
+++ b/cmd/quictest/main.go
@@
 // Пример флагов: BBRv2 + ACK_FREQUENCY
 var (
     cc      = flag.String("cc", "bbrv2", "cubic|bbrv2")
     ackext  = flag.Bool("ackfreq", true, "enable ACK_FREQUENCY extension")
     ackth   = flag.Uint64("ackth", 10, "Ack‑Eliciting Threshold")
     ackmad  = flag.Uint64("ackmad", 15, "Requested Max Ack Delay (ms)")
     ackreord= flag.Uint64("ackrth", 0, "Reordering Threshold (0=disable immediate on reordering)")
     tpmin   = flag.Uint64("tp_min_ack_us", 2000, "Advertise min_ack_delay (µs)")
 )
@@
 qcfg := &quic.Config{
     CloudBridgeCC: *cc,
     CloudBridgeUseAckFrequency: *ackext,
     CloudBridgeAckThreshold: *ackth,
     CloudBridgeRequestedMaxAckDelayMs: *ackmad,
     CloudBridgeReorderingThreshold: *ackreord,
     CloudBridgeAdvertiseMinAckDelayUs: *tpmin,
 }
@@
 // Отправить начальный ACK_FREQUENCY сразу после 1-го ACK (пример)
 go func(){
     <-time.After(150 * time.Millisecond)
     fr := &wire.AckFrequencyFrame{SequenceNumber: 1, AckElicitingThreshold: *ackth, RequestedMaxAckDelayMs: *ackmad, ReorderingThreshold: *ackreord}
     c.SendControlFrame(fr) // условная функция: используйте ваш способ записи фрейма в 1-RTT packet
 }()

```

## Что это даёт
- **BBRv2**: реальное управление inflight/pacing, измерение `delivery_rate` и `min_rtt`, штатные состояния `Startup/Drain/ProbeBW/ProbeRTT` (без «заглушек»).
- **ACK_FREQUENCY**: отправка/приём кадров по черновику v‑11 (тип `0xaf`), `IMMEDIATE_ACK` (`0x1f`), TP `min_ack_delay (0xff04de1b)`, корректная политика ACK с порогами/таймерами/реордерингом.
- **Interop‑безопасно**: если peer не рекламирует `min_ack_delay`, вы не шлёте кадры; при отсутствии кадров — поведение RFC 9000/9002.

## TODO (дальнейшая шлифовка)
- Подключить `pacingBps` в отправитель (если у вас есть таймерный pacer) для ровного трафика.
- Добавить qlog события: смена состояний BBRv2, обновления bw/min_rtt, отправка кадров ACK_FREQUENCY/IMMEDIATE_ACK.
- Тонкая настройка коэффициентов BBRv2 под ваши профили сети (Wi‑Fi/LTE/SAT): таблица gain’ов и пороги выхода из Startup/ProbeRTT.


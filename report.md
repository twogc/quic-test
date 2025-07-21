# 2GC CloudBridge QUICK testing

**Параметры:** "{Mode:client Addr:127.0.0.1:9000 Streams:2 Connections:2 Duration:0s PacketSize:800 Rate:100 ReportPath:report.md ReportFormat:md CertPath: KeyPath: Pattern:random NoTLS:false Prometheus:false EmulateLoss:0.05 EmulateLatency:10ms EmulateDup:0.01}"

**Метрики:** "&{mu:{_:{} mu:{state:0 sema:0}} Success:0 Errors:2 BytesSent:0 Latencies:[] Timestamps:[] Throughput:[] TimeSeriesLatency:[] TimeSeriesThroughput:[] PacketLoss:0 Retransmits:0 HandshakeTimes:[2.0724020000000003 1.734184] TLSVersion: CipherSuite: SessionResumptionCount:0 ZeroRTTCount:0 OneRTTCount:0 OutOfOrderCount:0 FlowControlEvents:0 KeyUpdateEvents:0 ErrorTypeCounts:map[quic_handshake:2] TimeSeriesPacketLoss:[] TimeSeriesRetransmits:[] TimeSeriesHandshakeTime:[{Time:0.002072687 Value:2.0724020000000003} {Time:0.001734469 Value:1.734184}]}"

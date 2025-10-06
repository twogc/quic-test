package client

import (
	"testing"
	"time"
)

func TestNewAdvancedPrometheusExporter(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	if exporter == nil {
		t.Fatal("NewAdvancedPrometheusExporter returned nil")
	}
	
	if exporter.metrics == nil {
		t.Error("metrics is nil")
	}
	
	if exporter.clientMetrics == nil {
		t.Error("clientMetrics is nil")
	}
	
	if exporter.testTypeCounters == nil {
		t.Error("testTypeCounters is nil")
	}
}

func TestUpdateTestType(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	exporter.UpdateTestType("latency", "random")
	
	metrics := exporter.GetClientMetrics()
	if metrics.TestType != "latency" {
		t.Errorf("Expected TestType 'latency', got '%s'", metrics.TestType)
	}
	if metrics.DataPattern != "random" {
		t.Errorf("Expected DataPattern 'random', got '%s'", metrics.DataPattern)
	}
}

func TestRecordTestExecution(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.RecordTestExecution("conn1", 100*time.Millisecond, "success")
}

func TestRecordConnectionInfo(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.RecordConnectionInfo("conn1", "127.0.0.1:9000", "TLS1.3", "AES256-GCM")
}

func TestRecordStreamInfo(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.RecordStreamInfo("stream1", "conn1", "bidirectional", "active")
}

func TestRecordLatency(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.RecordLatency(50 * time.Millisecond)
}

func TestRecordJitter(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.RecordJitter(5 * time.Millisecond)
}

func TestRecordThroughput(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.RecordThroughput(1024.0)
}

func TestRecordHandshakeTime(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.RecordHandshakeTime(200 * time.Millisecond)
}

func TestRecordRTT(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.RecordRTT(30 * time.Millisecond)
}

func TestIncrementConnections(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.IncrementConnections()
	exporter.DecrementConnections()
}

func TestIncrementStreams(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.IncrementStreams()
	exporter.DecrementStreams()
}

func TestAddBytes(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.AddBytesSent(1024)
	exporter.AddBytesReceived(2048)
}

func TestIncrementCounters(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.IncrementErrors()
	exporter.IncrementRetransmits()
	exporter.IncrementHandshakes()
	exporter.IncrementZeroRTT()
	exporter.IncrementOneRTT()
	exporter.IncrementSessionResumptions()
}

func TestSetGauges(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.SetCurrentThroughput(1024.0)
	exporter.SetCurrentLatency(50 * time.Millisecond)
	exporter.SetPacketLossRate(0.01)
	exporter.SetConnectionDuration(30 * time.Second)
}

func TestRecordEvents(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	// Не должно паниковать
	exporter.RecordScenarioEvent("test", "conn1", "stream1", "success")
	exporter.RecordErrorEvent("timeout", "conn1", "stream1", "warning")
	exporter.RecordProtocolEvent("handshake", "conn1", "TLS1.3", "AES256-GCM")
	exporter.RecordNetworkLatency("wifi", "conn1", "us-east", 50*time.Millisecond)
}

func TestGetClientMetrics(t *testing.T) {
	exporter := NewAdvancedPrometheusExporter()
	
	metrics := exporter.GetClientMetrics()
	if metrics == nil {
		t.Error("GetClientMetrics returned nil")
	}
	
	if metrics.StartTime.IsZero() {
		t.Error("StartTime is zero")
	}
}

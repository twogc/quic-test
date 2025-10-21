package main

import (
	"flag"
	"fmt"
	"log"

	"quic-test/internal"
)

func main() {
	var (
		tlsVersion     = flag.String("tls-version", "TLS 1.3", "TLS version (TLS 1.2, TLS 1.3)")
		cipherSuites   = flag.String("ciphers", "AES-256-GCM,ChaCha20-Poly1305", "Comma-separated cipher suites")
		certValidation = flag.Bool("cert-validation", true, "Enable certificate validation")
		enable0RTT     = flag.Bool("0rtt", false, "Enable 0-RTT")
		enableKeyUpdate = flag.Bool("key-update", true, "Enable key rotation")
		enableAntiReplay = flag.Bool("anti-replay", true, "Enable anti-replay protection")
		simulateAttacks = flag.Bool("simulate-attacks", false, "Simulate security attacks")
		attackTypes     = flag.String("attack-types", "MITM,Replay,DoS", "Comma-separated attack types")
		monitorCrypto   = flag.Bool("monitor-crypto", true, "Monitor cryptographic operations")
		monitorHandshake = flag.Bool("monitor-handshake", true, "Monitor handshake process")
		monitorTraffic  = flag.Bool("monitor-traffic", true, "Monitor traffic patterns")
		checkCompliance = flag.Bool("check-compliance", true, "Check security compliance")
		standards       = flag.String("standards", "RFC 9000,RFC 9001", "Comma-separated security standards")
	)
	flag.Parse()

	fmt.Println("🔒 QUIC Security Testing")
	fmt.Println("=======================")

	// Parse comma-separated values
	ciphers := parseCommaSeparated(*cipherSuites)
	attacks := parseCommaSeparated(*attackTypes)
	standardsList := parseCommaSeparated(*standards)

	// Create security test config
	config := internal.SecurityTestConfig{
		TLSVersion:       *tlsVersion,
		CipherSuites:     ciphers,
		CertValidation:   *certValidation,
		Enable0RTT:       *enable0RTT,
		EnableKeyUpdate:  *enableKeyUpdate,
		EnableAntiReplay: *enableAntiReplay,
		SimulateAttacks:  *simulateAttacks,
		AttackTypes:      attacks,
		MonitorCrypto:    *monitorCrypto,
		MonitorHandshake: *monitorHandshake,
		MonitorTraffic:   *monitorTraffic,
		CheckCompliance: *checkCompliance,
		Standards:        standardsList,
	}

	// Create security tester
	tester := internal.NewSecurityTester(config)

	// Run security tests
	if err := tester.RunSecurityTests(); err != nil {
		log.Fatalf("Security testing failed: %v", err)
	}

	// Display results
	results := tester.GetResults()
	overallScore := tester.GetOverallScore()

	fmt.Printf("\n📊 Security Test Results:\n")
	fmt.Printf("========================\n")
	fmt.Printf("Overall Security Score: %.2f%%\n", overallScore*100)
	fmt.Printf("Total Tests: %d\n", len(results))

	for _, result := range results {
		status := "❌ FAILED"
		if result.Passed {
			status = "✅ PASSED"
		}
		fmt.Printf("\n%s %s (Score: %.2f)\n", status, result.TestName, result.Score)

		if len(result.Vulnerabilities) > 0 {
			fmt.Printf("  Vulnerabilities found:\n")
			for _, vuln := range result.Vulnerabilities {
				fmt.Printf("    ⚠️  [%s] %s: %s (CVSS: %.1f)\n", 
					vuln.Severity, vuln.Type, vuln.Description, vuln.CVSS)
				if vuln.Mitigation != "" {
					fmt.Printf("      💡 Mitigation: %s\n", vuln.Mitigation)
				}
			}
		}

		if len(result.Recommendations) > 0 {
			fmt.Printf("  Recommendations:\n")
			for _, rec := range result.Recommendations {
				fmt.Printf("    💡 %s\n", rec)
			}
		}
	}

	// Security summary
	fmt.Printf("\n🔐 Security Summary:\n")
	fmt.Printf("==================\n")
	
	totalVulns := 0
	criticalVulns := 0
	highVulns := 0
	
	for _, result := range results {
		totalVulns += len(result.Vulnerabilities)
		for _, vuln := range result.Vulnerabilities {
			switch vuln.Severity {
			case "Critical":
				criticalVulns++
			case "High":
				highVulns++
			}
		}
	}

	fmt.Printf("Total Vulnerabilities: %d\n", totalVulns)
	fmt.Printf("Critical: %d\n", criticalVulns)
	fmt.Printf("High: %d\n", highVulns)

	if overallScore >= 0.8 {
		fmt.Printf("🟢 Security Status: GOOD\n")
	} else if overallScore >= 0.6 {
		fmt.Printf("🟡 Security Status: FAIR\n")
	} else {
		fmt.Printf("🔴 Security Status: POOR\n")
	}

	fmt.Println("\n✅ Security testing completed")
}

func parseCommaSeparated(input string) []string {
	if input == "" {
		return []string{}
	}
	
	var result []string
	for _, item := range splitComma(input) {
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func splitComma(input string) []string {
	// Simple comma splitting (could be improved with proper CSV parsing)
	var result []string
	var current string
	
	for _, char := range input {
		if char == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		result = append(result, current)
	}
	
	return result
}

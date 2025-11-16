#!/usr/bin/env python3

"""
BBRv2 vs BBRv3 Comparison Analysis
Compares Phase 0a (BBRv2) and Phase 0b (BBRv3) baseline test results
to demonstrate BBRv3 optimizations and improvements
"""

import json
import sys
from pathlib import Path
from typing import Dict, List, Any
import statistics

class BBRComparison:
    def __init__(self, bbrv2_dir: str, bbrv3_dir: str):
        self.bbrv2_dir = Path(bbrv2_dir)
        self.bbrv3_dir = Path(bbrv3_dir)
        self.bbrv2_tests = {}
        self.bbrv3_tests = {}
        self._load_tests()

    def _load_tests(self):
        """Load all JSON test reports"""
        print(f"🔍 Loading BBRv2 tests from {self.bbrv2_dir}...\n")
        json_files = sorted(self.bbrv2_dir.glob('*.json'))
        for json_file in json_files:
            try:
                with open(json_file) as f:
                    data = json.load(f)
                self.bbrv2_tests[json_file.stem] = data
                print(f"✅ Loaded: {json_file.name}")
            except Exception as e:
                print(f"❌ Error loading {json_file.name}: {e}")

        print(f"\n🔍 Loading BBRv3 tests from {self.bbrv3_dir}...\n")
        json_files = sorted(self.bbrv3_dir.glob('*.json'))
        for json_file in json_files:
            try:
                with open(json_file) as f:
                    data = json.load(f)
                self.bbrv3_tests[json_file.stem] = data
                print(f"✅ Loaded: {json_file.name}")
            except Exception as e:
                print(f"❌ Error loading {json_file.name}: {e}")

        print(f"\nTotal BBRv2 tests loaded: {len(self.bbrv2_tests)}")
        print(f"Total BBRv3 tests loaded: {len(self.bbrv3_tests)}\n")

    def extract_metrics(self, test_data: Dict) -> Dict:
        """Extract critical metrics from test result"""
        try:
            metrics = test_data.get('metrics', {})
            latency = metrics.get('latency', {})
            config = test_data.get('test_config', {})

            # Calculate throughput in Mbps
            duration_s = config.get('duration', 60) / 1e9 if isinstance(config.get('duration'), (int, float)) else 60
            throughput_mbps = (metrics.get('bytes_sent', 0) * 8) / (duration_s * 1e6) if duration_s > 0 else 0

            # Calculate bufferbloat factor
            bufferbloat = 0
            if latency.get('min', 0) > 0:
                bufferbloat = (latency.get('average', 0) / latency.get('min', 1)) - 1

            return {
                'throughput_mbps': throughput_mbps,
                'jitter_ms': latency.get('jitter', 0) * 1e3,
                'avg_latency_ms': latency.get('average', 0) * 1e3,
                'p50_latency_ms': latency.get('p50', 0) * 1e3,
                'p95_latency_ms': latency.get('p95', 0) * 1e3,
                'p99_latency_ms': latency.get('p99', 0) * 1e3,
                'bufferbloat_factor': bufferbloat,
                'packet_loss_pct': metrics.get('packet_loss', 0),
                'retransmits': metrics.get('retransmits', 0),
                'bytes_sent': metrics.get('bytes_sent', 0),
                'connections': config.get('connections', 0),
                'streams': config.get('streams', 0),
            }
        except Exception as e:
            print(f"⚠️  Error extracting metrics: {e}")
            return {}

    def print_comparison_table(self):
        """Print detailed comparison table"""
        print("\n" + "="*200)
        print("  BBRv2 vs BBRv3 DETAILED COMPARISON")
        print("="*200 + "\n")

        print(f"{'Profile':<12} {'Load':<7} {'Metric':<20} {'BBRv2':<20} {'BBRv3':<20} {'Change':<15}")
        print("-" * 200)

        profiles = ['good', 'mobile', 'satellite']
        for profile in profiles:
            for load in ['light', 'medium']:
                bbrv2_name = f"baseline_bbrv2_{profile}_{load}"
                bbrv3_name = f"baseline_bbrv3_{profile}_{load}"

                if bbrv2_name in self.bbrv2_tests and bbrv3_name in self.bbrv3_tests:
                    v2_metrics = self.extract_metrics(self.bbrv2_tests[bbrv2_name])
                    v3_metrics = self.extract_metrics(self.bbrv3_tests[bbrv3_name])

                    # Throughput comparison
                    v2_tp = v2_metrics.get('throughput_mbps', 0)
                    v3_tp = v3_metrics.get('throughput_mbps', 0)
                    tp_change = ((v3_tp / v2_tp - 1) * 100) if v2_tp > 0 else 0
                    print(f"{profile:<12} {load:<7} {'Throughput (Mbps)':<20} {v2_tp:>19.2f} {v3_tp:>19.2f} {tp_change:>13.1f}%")

                    # Jitter comparison
                    v2_jitter = v2_metrics.get('jitter_ms', 0)
                    v3_jitter = v3_metrics.get('jitter_ms', 0)
                    jitter_change = ((v3_jitter / v2_jitter - 1) * 100) if v2_jitter > 0 else 0
                    print(f"{'':12} {'':7} {'Jitter (ms)':<20} {v2_jitter:>19.1f} {v3_jitter:>19.1f} {jitter_change:>13.1f}%")

                    # Avg latency comparison
                    v2_lat = v2_metrics.get('avg_latency_ms', 0)
                    v3_lat = v3_metrics.get('avg_latency_ms', 0)
                    lat_change = ((v3_lat / v2_lat - 1) * 100) if v2_lat > 0 else 0
                    print(f"{'':12} {'':7} {'Avg Latency (ms)':<20} {v2_lat:>19.1f} {v3_lat:>19.1f} {lat_change:>13.1f}%")

                    # P95 latency comparison
                    v2_p95 = v2_metrics.get('p95_latency_ms', 0)
                    v3_p95 = v3_metrics.get('p95_latency_ms', 0)
                    p95_change = ((v3_p95 / v2_p95 - 1) * 100) if v2_p95 > 0 else 0
                    print(f"{'':12} {'':7} {'P95 Latency (ms)':<20} {v2_p95:>19.1f} {v3_p95:>19.1f} {p95_change:>13.1f}%")

                    # Bufferbloat comparison
                    v2_bb = v2_metrics.get('bufferbloat_factor', 0)
                    v3_bb = v3_metrics.get('bufferbloat_factor', 0)
                    bb_change = ((v3_bb / v2_bb - 1) * 100) if v2_bb > 0 else 0
                    print(f"{'':12} {'':7} {'Bufferbloat (×)':<20} {v2_bb:>19.2f} {v3_bb:>19.2f} {bb_change:>13.1f}%")

                    print()

    def print_key_improvements(self):
        """Highlight key improvements"""
        print("\n" + "="*180)
        print("  KEY BBRv3 IMPROVEMENTS OVER BBRv2")
        print("="*180 + "\n")

        improvements = {
            'throughput': [],
            'jitter': [],
            'latency': [],
            'bufferbloat': []
        }

        profiles = ['good', 'mobile', 'satellite']
        for profile in profiles:
            for load in ['light', 'medium']:
                bbrv2_name = f"baseline_bbrv2_{profile}_{load}"
                bbrv3_name = f"baseline_bbrv3_{profile}_{load}"

                if bbrv2_name in self.bbrv2_tests and bbrv3_name in self.bbrv3_tests:
                    v2_metrics = self.extract_metrics(self.bbrv2_tests[bbrv2_name])
                    v3_metrics = self.extract_metrics(self.bbrv3_tests[bbrv3_name])

                    # Track improvements
                    tp_improvement = ((v3_metrics.get('throughput_mbps', 0) / (v2_metrics.get('throughput_mbps', 1))) - 1) * 100
                    jitter_improvement = ((v2_metrics.get('jitter_ms', 1) / (v3_metrics.get('jitter_ms', 1))) - 1) * 100
                    latency_improvement = ((v2_metrics.get('avg_latency_ms', 1) / (v3_metrics.get('avg_latency_ms', 1))) - 1) * 100
                    bufferbloat_improvement = ((v2_metrics.get('bufferbloat_factor', 1) / (v3_metrics.get('bufferbloat_factor', 1))) - 1) * 100

                    improvements['throughput'].append((f"{profile}_{load}", tp_improvement))
                    improvements['jitter'].append((f"{profile}_{load}", jitter_improvement))
                    improvements['latency'].append((f"{profile}_{load}", latency_improvement))
                    improvements['bufferbloat'].append((f"{profile}_{load}", bufferbloat_improvement))

        # Print throughput improvements
        print("📈 THROUGHPUT IMPROVEMENT (higher is better):")
        for name, improvement in improvements['throughput']:
            status = "✅" if improvement > 0 else "⚠️"
            print(f"   {status} {name:<20}: {improvement:>+7.1f}%")
        print()

        # Print jitter improvements
        print("📉 JITTER REDUCTION (higher is better - lower is better):")
        for name, improvement in improvements['jitter']:
            status = "✅" if improvement > 0 else "⚠️"
            print(f"   {status} {name:<20}: {improvement:>+7.1f}% reduction")
        print()

        # Print latency improvements
        print("⏱️  LATENCY REDUCTION (higher is better):")
        for name, improvement in improvements['latency']:
            status = "✅" if improvement > 0 else "⚠️"
            print(f"   {status} {name:<20}: {improvement:>+7.1f}% reduction")
        print()

        # Print bufferbloat improvements
        print("BUFFERBLOAT REDUCTION (higher is better):")
        for name, improvement in improvements['bufferbloat']:
            status = "✅" if improvement > 0 else "⚠️"
            print(f"   {status} {name:<20}: {improvement:>+7.1f}% reduction")

    def print_profile_summary(self):
        """Print per-profile summary"""
        print("\n" + "="*150)
        print("  PROFILE-LEVEL SUMMARY")
        print("="*150 + "\n")

        profiles = ['good', 'mobile', 'satellite']
        for profile in profiles:
            print(f"📡 {profile.upper()} Profile:")
            print("─" * 80)

            light_v2 = self.extract_metrics(self.bbrv2_tests.get(f"baseline_bbrv2_{profile}_light", {}))
            light_v3 = self.extract_metrics(self.bbrv3_tests.get(f"baseline_bbrv3_{profile}_light", {}))
            medium_v2 = self.extract_metrics(self.bbrv2_tests.get(f"baseline_bbrv2_{profile}_medium", {}))
            medium_v3 = self.extract_metrics(self.bbrv3_tests.get(f"baseline_bbrv3_{profile}_medium", {}))

            if light_v2 and light_v3:
                tp_improvement = ((light_v3['throughput_mbps'] / light_v2['throughput_mbps']) - 1) * 100
                jitter_improvement = ((light_v2['jitter_ms'] / light_v3['jitter_ms']) - 1) * 100
                print(f"  Light Load:")
                print(f"    • Throughput: {light_v2['throughput_mbps']:.2f} Mbps (v2) → {light_v3['throughput_mbps']:.2f} Mbps (v3) [{tp_improvement:+.1f}%]")
                print(f"    • Jitter: {light_v2['jitter_ms']:.1f} ms (v2) → {light_v3['jitter_ms']:.1f} ms (v3) [{jitter_improvement:+.1f}% reduction]")

            if medium_v2 and medium_v3:
                tp_improvement = ((medium_v3['throughput_mbps'] / medium_v2['throughput_mbps']) - 1) * 100
                jitter_improvement = ((medium_v2['jitter_ms'] / medium_v3['jitter_ms']) - 1) * 100
                print(f"  Medium Load:")
                print(f"    • Throughput: {medium_v2['throughput_mbps']:.2f} Mbps (v2) → {medium_v3['throughput_mbps']:.2f} Mbps (v3) [{tp_improvement:+.1f}%]")
                print(f"    • Jitter: {medium_v2['jitter_ms']:.1f} ms (v2) → {medium_v3['jitter_ms']:.1f} ms (v3) [{jitter_improvement:+.1f}% reduction]")
            print()

def main():
    if len(sys.argv) < 3:
        print("Usage: python3 compare_bbrv2_vs_bbrv3.py <bbrv2_dir> <bbrv3_dir>")
        sys.exit(1)

    bbrv2_dir = sys.argv[1]
    bbrv3_dir = sys.argv[2]

    comparison = BBRComparison(bbrv2_dir, bbrv3_dir)

    print("\n" + "="*200)
    print("  BBRv2 vs BBRv3 COMPREHENSIVE ANALYSIS")
    print("  Demonstrating BBRv3 Optimizations and Improvements")
    print("="*200 + "\n")

    comparison.print_comparison_table()
    comparison.print_key_improvements()
    comparison.print_profile_summary()

    print("\n✅ BBRv2 vs BBRv3 analysis complete!")
    print(f"BBRv2 results: {bbrv2_dir}")
    print(f"BBRv3 results: {bbrv3_dir}")
    print()

if __name__ == '__main__':
    main()

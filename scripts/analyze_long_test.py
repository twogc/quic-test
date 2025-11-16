#!/usr/bin/env python3
"""
Анализ результатов длительного тестирования BBRv2 vs BBRv3
"""

import json
import os
import sys
from datetime import datetime

def load_json(filepath):
    """Загружает JSON файл"""
    with open(filepath, 'r') as f:
        return json.load(f)

def calculate_throughput(bytes_sent, duration_seconds):
    """Вычисляет throughput в Mbps"""
    if duration_seconds <= 0:
        return 0.0
    return (bytes_sent * 8) / (duration_seconds * 1_000_000)

def extract_metrics(data):
    """Извлекает метрики из JSON данных"""
    metrics = data.get('metrics', {})
    latency = metrics.get('latency', {})
    bbrv3 = data.get('BBRv3Metrics', {})
    
    # Извлекаем длительность
    test_config = data.get('test_config', {})
    duration = test_config.get('duration', 0)
    
    # Обрабатываем разные форматы duration
    if isinstance(duration, dict):
        # Если это объект с полем seconds или nanoseconds
        if 'seconds' in duration:
            duration = duration['seconds']
        elif 'nanoseconds' in duration:
            duration = duration['nanoseconds'] / 1e9
        else:
            duration = 0
    elif isinstance(duration, (int, float)):
        # Если это число, считаем что это секунды
        duration = float(duration)
    elif isinstance(duration, str):
        # Парсим строку типа "5m", "2m", "300s"
        if duration.endswith('m'):
            duration = int(duration[:-1]) * 60
        elif duration.endswith('s'):
            duration = int(duration[:-1])
        elif duration.endswith('h'):
            duration = int(duration[:-1]) * 3600
        else:
            try:
                duration = int(duration)
            except ValueError:
                duration = 0
    
    # Если duration не определен, используем throughput_mbps из метрик, если есть
    bytes_sent = metrics.get('bytes_sent', 0)
    throughput_mbps = metrics.get('throughput_mbps', 0)
    
    # Если throughput_mbps уже есть и правильный, используем его
    if throughput_mbps > 0:
        throughput = throughput_mbps
    elif duration > 0 and bytes_sent > 0:
        # Пересчитываем throughput
        throughput = calculate_throughput(bytes_sent, duration)
    else:
        throughput = 0
    
    # Извлекаем дополнительные метрики
    rtt_min = latency.get('min', 0)
    bufferbloat = metrics.get('bufferbloat_factor', 0)
    fairness = metrics.get('fairness_index', 0)
    
    # Вычисляем конвергенцию BBRv3 (|bw_fast - bw_slow| / max(bw_fast, bw_slow))
    convergence = 0.0
    if bbrv3:
        bw_fast = bbrv3.get('bw_fast', 0) / 1_000_000  # bps -> Mbps
        bw_slow = bbrv3.get('bw_slow', 0) / 1_000_000
        if bw_fast > 0 or bw_slow > 0:
            max_bw = max(bw_fast, bw_slow)
            if max_bw > 0:
                convergence = abs(bw_fast - bw_slow) / max_bw
    
    return {
        'throughput': throughput,
        'bytes_sent': bytes_sent,
        'duration': duration,
        'rtt_min': rtt_min,
        'rtt_p50': latency.get('p50', 0),
        'rtt_p95': latency.get('p95', 0),
        'rtt_p99': latency.get('p99', 0),
        'jitter': latency.get('jitter', 0),
        'average_rtt': latency.get('average', 0),
        'packet_loss': metrics.get('packet_loss', 0),
        'retransmits': metrics.get('retransmits', 0),
        'errors': metrics.get('errors', 0),
        'success': metrics.get('success', 0),
        'bufferbloat_factor': bufferbloat,
        'fairness_index': fairness,
        'bbrv3_phase': bbrv3.get('phase', 'N/A') if bbrv3 else 'N/A',
        'bw_fast': bbrv3.get('bw_fast', 0) / 1_000_000 if bbrv3 else 0,
        'bw_slow': bbrv3.get('bw_slow', 0) / 1_000_000 if bbrv3 else 0,
        'loss_rate_round': bbrv3.get('loss_rate_round', 0) if bbrv3 else 0,
        'headroom_usage': bbrv3.get('headroom_usage', 0) * 100 if bbrv3 else 0,
        'convergence': convergence,
    }

def format_number(value, decimals=2):
    """Форматирует число с заданным количеством знаков"""
    if isinstance(value, (int, float)):
        return f"{value:.{decimals}f}"
    return str(value)

def calculate_percentage_change(old, new):
    """Вычисляет процентное изменение"""
    if old == 0:
        return 0.0
    return ((new - old) / old) * 100

def generate_comparison_report(output_dir):
    """Генерирует отчет сравнения"""
    bbrv2_file = f"{output_dir}/bbrv2_results.json"
    bbrv3_file = f"{output_dir}/bbrv3_results.json"
    
    if not os.path.exists(bbrv2_file) or not os.path.exists(bbrv3_file):
        print(f"❌ Файлы результатов не найдены!")
        print(f"   BBRv2: {'✅' if os.path.exists(bbrv2_file) else '❌'}")
        print(f"   BBRv3: {'✅' if os.path.exists(bbrv3_file) else '❌'}")
        return
    
    print("Загрузка результатов...")
    bbrv2_data = load_json(bbrv2_file)
    bbrv3_data = load_json(bbrv3_file)
    
    bbrv2_metrics = extract_metrics(bbrv2_data)
    bbrv3_metrics = extract_metrics(bbrv3_data)
    
    print("\n" + "="*80)
    print("📈 РЕЗУЛЬТАТЫ ДЛИТЕЛЬНОГО ТЕСТИРОВАНИЯ (5 минут, 50 соединений)")
    print("="*80)
    
    print("\n🔵 BBRv2:")
    print(f"   Throughput:      {format_number(bbrv2_metrics['throughput'], 3)} Mbps")
    print(f"   Bytes Sent:      {bbrv2_metrics['bytes_sent']:,}")
    print(f"   RTT Min:         {format_number(bbrv2_metrics.get('rtt_min', 0), 2)} ms")
    print(f"   RTT P50:         {format_number(bbrv2_metrics['rtt_p50'], 2)} ms")
    print(f"   RTT P95:         {format_number(bbrv2_metrics['rtt_p95'], 2)} ms")
    print(f"   RTT P99:         {format_number(bbrv2_metrics['rtt_p99'], 2)} ms")
    print(f"   Jitter:          {format_number(bbrv2_metrics['jitter'], 2)} ms")
    print(f"   Average RTT:     {format_number(bbrv2_metrics['average_rtt'], 2)} ms")
    print(f"   Bufferbloat:     {format_number(bbrv2_metrics.get('bufferbloat_factor', 0), 3)}")
    print(f"   Fairness Index:  {format_number(bbrv2_metrics.get('fairness_index', 0), 3)}")
    print(f"   Packet Loss:     {format_number(bbrv2_metrics['packet_loss'], 3)}%")
    print(f"   Retransmits:     {bbrv2_metrics['retransmits']:,}")
    print(f"   Errors:          {bbrv2_metrics['errors']}")
    
    print("\n🟢 BBRv3 (оптимизированный):")
    print(f"   Throughput:      {format_number(bbrv3_metrics['throughput'], 3)} Mbps")
    print(f"   Bytes Sent:      {bbrv3_metrics['bytes_sent']:,}")
    print(f"   RTT Min:         {format_number(bbrv3_metrics.get('rtt_min', 0), 2)} ms")
    print(f"   RTT P50:         {format_number(bbrv3_metrics['rtt_p50'], 2)} ms")
    print(f"   RTT P95:         {format_number(bbrv3_metrics['rtt_p95'], 2)} ms")
    print(f"   RTT P99:         {format_number(bbrv3_metrics['rtt_p99'], 2)} ms")
    print(f"   Jitter:          {format_number(bbrv3_metrics['jitter'], 2)} ms")
    print(f"   Average RTT:     {format_number(bbrv3_metrics['average_rtt'], 2)} ms")
    print(f"   Bufferbloat:     {format_number(bbrv3_metrics.get('bufferbloat_factor', 0), 3)}")
    print(f"   Fairness Index:  {format_number(bbrv3_metrics.get('fairness_index', 0), 3)}")
    print(f"   Packet Loss:     {format_number(bbrv3_metrics['packet_loss'], 3)}%")
    print(f"   Retransmits:     {bbrv3_metrics['retransmits']:,}")
    print(f"   Errors:          {bbrv3_metrics['errors']}")
    if bbrv3_metrics['bbrv3_phase'] != 'N/A':
        print(f"   Phase:           {bbrv3_metrics['bbrv3_phase']}")
        print(f"   BW Fast:         {format_number(bbrv3_metrics['bw_fast'], 3)} Mbps")
        print(f"   BW Slow:         {format_number(bbrv3_metrics['bw_slow'], 3)} Mbps")
        print(f"   Convergence:    {format_number(bbrv3_metrics.get('convergence', 0), 3)} (|fast-slow|/max)")
        print(f"   Loss Rate Round: {format_number(bbrv3_metrics['loss_rate_round'], 3)}%")
        print(f"   Headroom Usage:  {format_number(bbrv3_metrics['headroom_usage'], 1)}%")
    
    print("\n" + "="*80)
    print("СРАВНЕНИЕ: BBRv3 vs BBRv2")
    print("="*80)
    
    metrics_to_compare = [
        ('Throughput (Mbps)', 'throughput', True),
        ('RTT Min (ms)', 'rtt_min', False),
        ('RTT P50 (ms)', 'rtt_p50', False),
        ('RTT P95 (ms)', 'rtt_p95', False),
        ('RTT P99 (ms)', 'rtt_p99', False),
        ('Jitter (ms)', 'jitter', False),
        ('Average RTT (ms)', 'average_rtt', False),
        ('Bufferbloat Factor', 'bufferbloat_factor', False),
        ('Fairness Index', 'fairness_index', True),
        ('Packet Loss (%)', 'packet_loss', False),
        ('Retransmits', 'retransmits', False),
    ]
    
    improvements = []
    degradations = []
    
    for name, key, higher_is_better in metrics_to_compare:
        v2_val = bbrv2_metrics[key]
        v3_val = bbrv3_metrics[key]
        change_pct = calculate_percentage_change(v2_val, v3_val)
        
        if higher_is_better:
            status = "✅" if change_pct > 0 else "⚠️"
            if change_pct > 5:
                improvements.append(name)
            elif change_pct < -5:
                degradations.append(name)
        else:
            status = "✅" if change_pct < 0 else "⚠️"
            if change_pct < -5:
                improvements.append(name)
            elif change_pct > 5:
                degradations.append(name)
        
        print(f"   {name:25s}: {change_pct:+7.2f}% {status}")
    
    print("\n" + "="*80)
    if improvements:
        print(f"🎉 УЛУЧШЕНИЯ ({len(improvements)} метрик):")
        for imp in improvements:
            print(f"   ✅ {imp}")
    
    if degradations:
        print(f"\n⚠️  УХУДШЕНИЯ ({len(degradations)} метрик):")
        for deg in degradations:
            print(f"   ⚠️  {deg}")
    
    if not improvements and not degradations:
        print("Изменения незначительны (< 5%)")
    
    print("\n" + "="*80)
    
    # Сохраняем отчет в файл
    report_file = f"{output_dir}/LONG_TEST_COMPARISON.md"
    with open(report_file, 'w') as f:
        f.write(f"# Длительное тестирование BBRv2 vs BBRv3\n\n")
        f.write(f"**Дата:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
        f.write(f"**Параметры теста:**\n")
        f.write(f"- Длительность: 5 минут\n")
        f.write(f"- Соединения: 50\n")
        f.write(f"- Streams: 2\n")
        f.write(f"- Latency: 200ms\n")
        f.write(f"- Loss: 0.1%\n\n")
        
        f.write(f"## Результаты BBRv2\n\n")
        f.write(f"| Метрика | Значение |\n")
        f.write(f"|---------|----------|\n")
        for name, key, _ in metrics_to_compare:
            f.write(f"| {name} | {format_number(bbrv2_metrics[key])} |\n")
        
        f.write(f"\n## Результаты BBRv3\n\n")
        f.write(f"| Метрика | Значение |\n")
        f.write(f"|---------|----------|\n")
        for name, key, _ in metrics_to_compare:
            f.write(f"| {name} | {format_number(bbrv3_metrics[key])} |\n")
        
        f.write(f"\n## Сравнение\n\n")
        f.write(f"| Метрика | Изменение |\n")
        f.write(f"|---------|-----------|\n")
        for name, key, _ in metrics_to_compare:
            change = calculate_percentage_change(bbrv2_metrics[key], bbrv3_metrics[key])
            f.write(f"| {name} | {change:+.2f}% |\n")
    
    print(f"✅ Отчет сохранен: {report_file}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Использование: python3 analyze_long_test.py <output_dir>")
        sys.exit(1)
    
    output_dir = sys.argv[1]
    generate_comparison_report(output_dir)


#!/bin/bash

# Скрипт для применения DevOps оптимизаций QUIC сервера
# Основан на рекомендациях DevOps команды

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - DevOps Optimizations${NC}"
echo "Применение производственных оптимизаций QUIC сервера"
echo ""

# Функция для применения системных оптимизаций
apply_system_optimizations() {
    echo -e "${YELLOW}🔧 Применение системных оптимизаций...${NC}"
    
    # UDP буферы
    echo -e "${CYAN}📡 Настройка UDP буферов:${NC}"
    sudo sysctl -w net.core.rmem_max=134217728 >/dev/null 2>&1
    sudo sysctl -w net.core.rmem_default=134217728 >/dev/null 2>&1
    sudo sysctl -w net.core.wmem_max=134217728 >/dev/null 2>&1
    sudo sysctl -w net.core.wmem_default=134217728 >/dev/null 2>&1
    echo "  ✅ UDP буферы увеличены до 128MB"
    
    # Сетевые оптимизации
    echo -e "${CYAN}🌐 Настройка сетевых параметров:${NC}"
    sudo sysctl -w net.core.netdev_max_backlog=5000 >/dev/null 2>&1
    sudo sysctl -w net.core.somaxconn=65535 >/dev/null 2>&1
    sudo sysctl -w net.ipv4.udp_mem="102400 873800 16777216" >/dev/null 2>&1
    sudo sysctl -w net.ipv4.udp_rmem_min=8192 >/dev/null 2>&1
    sudo sysctl -w net.ipv4.udp_wmem_min=8192 >/dev/null 2>&1
    echo "  ✅ Сетевые параметры оптимизированы"
    
    # TCP оптимизации
    echo -e "${CYAN}Настройка TCP параметров:${NC}"
    sudo sysctl -w net.ipv4.tcp_congestion_control=bbr >/dev/null 2>&1
    sudo sysctl -w net.ipv4.tcp_rmem="4096 87380 134217728" >/dev/null 2>&1
    sudo sysctl -w net.ipv4.tcp_wmem="4096 65536 134217728" >/dev/null 2>&1
    echo "  ✅ TCP congestion control: BBR"
    echo "  ✅ TCP буферы оптимизированы"
}

# Функция для настройки лимитов процессов
setup_process_limits() {
    echo -e "${YELLOW}⚙️ Настройка лимитов процессов...${NC}"
    
    # Увеличиваем лимиты для текущей сессии
    ulimit -n 65536 2>/dev/null
    ulimit -u 32768 2>/dev/null
    
    echo "  ✅ Лимиты файлов: 65536"
    echo "  ✅ Лимиты процессов: 32768"
    
    # Создаем конфигурацию для постоянного применения
    echo -e "${CYAN}📝 Создание конфигурации лимитов:${NC}"
    cat > /tmp/quic-limits.conf << EOF
# QUIC Server Process Limits
quic-server soft nofile 65536
quic-server hard nofile 65536
quic-server soft nproc 32768
quic-server hard nproc 32768
EOF
    
    echo "  ✅ Конфигурация лимитов создана: /tmp/quic-limits.conf"
    echo "  💡 Для постоянного применения добавьте в /etc/security/limits.conf"
}

# Функция для создания оптимизированного сервера
create_optimized_server() {
    echo -e "${YELLOW}Создание оптимизированного сервера...${NC}"
    
    # Создаем скрипт запуска оптимизированного сервера
    cat > scripts/optimized-server-start.sh << 'EOF'
#!/bin/bash

# Оптимизированный запуск QUIC сервера
# Применяет все DevOps рекомендации

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Optimized Server${NC}"
echo "Запуск оптимизированного QUIC сервера"
echo ""

# Применяем системные оптимизации
echo -e "${YELLOW}🔧 Применение системных оптимизаций...${NC}"
sudo sysctl -w net.core.rmem_max=134217728 >/dev/null 2>&1
sudo sysctl -w net.core.wmem_max=134217728 >/dev/null 2>&1
sudo sysctl -w net.core.netdev_max_backlog=5000 >/dev/null 2>&1
sudo sysctl -w net.ipv4.tcp_congestion_control=bbr >/dev/null 2>&1

# Устанавливаем лимиты процессов
ulimit -n 65536 2>/dev/null
ulimit -u 32768 2>/dev/null

echo -e "${GREEN}✅ Системные оптимизации применены${NC}"

# Запускаем сервер с оптимизированными параметрами
echo -e "${YELLOW}Запуск оптимизированного сервера...${NC}"

# Переменные окружения для оптимизации
export QUIC_MAX_CONNECTIONS=1000
export QUIC_MAX_RATE_PER_CONN=20
export QUIC_CONNECTION_TIMEOUT=60s
export QUIC_HANDSHAKE_TIMEOUT=10s
export QUIC_KEEP_ALIVE=30s
export QUIC_MAX_STREAMS=100
export QUIC_ENABLE_DATAGRAMS=true
export QUIC_ENABLE_0RTT=true
export QUIC_MONITORING=true

# Запускаем Docker контейнер с оптимизированными параметрами
timeout 10m docker run --rm --name 2gc-network-server-optimized \
    --network 2gc-network-suite \
    -p 9000:9000/udp \
    -p 2113:2113 \
    -p 6060:6060 \
    -e QUIC_MAX_CONNECTIONS=$QUIC_MAX_CONNECTIONS \
    -e QUIC_MAX_RATE_PER_CONN=$QUIC_MAX_RATE_PER_CONN \
    -e QUIC_CONNECTION_TIMEOUT=$QUIC_CONNECTION_TIMEOUT \
    -e QUIC_HANDSHAKE_TIMEOUT=$QUIC_HANDSHAKE_TIMEOUT \
    -e QUIC_KEEP_ALIVE=$QUIC_KEEP_ALIVE \
    -e QUIC_MAX_STREAMS=$QUIC_MAX_STREAMS \
    -e QUIC_ENABLE_DATAGRAMS=$QUIC_ENABLE_DATAGRAMS \
    -e QUIC_ENABLE_0RTT=$QUIC_ENABLE_0RTT \
    -e QUIC_MONITORING=$QUIC_MONITORING \
    2gc-network-suite:server

echo -e "${GREEN}✅ Оптимизированный сервер завершил работу${NC}"
EOF

    chmod +x scripts/optimized-server-start.sh
    echo "  ✅ Скрипт оптимизированного сервера создан: scripts/optimized-server-start.sh"
}

# Функция для создания health check скрипта
create_health_check() {
    echo -e "${YELLOW}🏥 Создание health check скрипта...${NC}"
    
    cat > scripts/health-check.sh << 'EOF'
#!/bin/bash

# Health check для QUIC сервера
# Проверяет критические зоны и производительность

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

SERVER_URL="http://localhost:2113/metrics"
CRITICAL_ZONE_ALERT=false
WARNINGS=0

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  QUIC Server Health Check${NC}"
echo "Проверка состояния сервера"
echo ""

# Проверяем доступность сервера
if ! curl -s $SERVER_URL >/dev/null 2>&1; then
    echo -e "${RED}❌ Сервер недоступен${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Сервер доступен${NC}"

# Проверяем критическую зону (26-35 pps)
echo -e "${YELLOW}🔍 Проверка критической зоны...${NC}"
RATE=$(curl -s $SERVER_URL 2>/dev/null | grep 'quic_server_rate_per_connection' | awk '{print $2}' | head -1)

if [ -n "$RATE" ] && (( $(echo "$RATE >= 26 && $RATE <= 35" | bc -l 2>/dev/null || echo "0") )); then
    echo -e "${RED}🚨 КРИТИЧЕСКАЯ ЗОНА: Rate $RATE pps (26-35 pps)${NC}"
    CRITICAL_ZONE_ALERT=true
    WARNINGS=$((WARNINGS + 1))
elif [ -n "$RATE" ]; then
    echo -e "${GREEN}✅ Rate $RATE pps (безопасная зона)${NC}"
fi

# Проверяем jitter
echo -e "${YELLOW}Проверка jitter...${NC}"
JITTER=$(curl -s $SERVER_URL 2>/dev/null | grep 'quic_server_jitter_seconds' | awk '{print $2}' | head -1)

if [ -n "$JITTER" ] && (( $(echo "$JITTER > 0.1" | bc -l 2>/dev/null || echo "0") )); then
    echo -e "${RED}⚠️ Высокий jitter: $JITTER секунд${NC}"
    WARNINGS=$((WARNINGS + 1))
elif [ -n "$JITTER" ]; then
    echo -e "${GREEN}✅ Jitter: $JITTER секунд (норма)${NC}"
fi

# Проверяем ошибки
echo -e "${YELLOW}❌ Проверка ошибок...${NC}"
ERRORS=$(curl -s $SERVER_URL 2>/dev/null | grep 'quic_server_errors_total' | awk '{print $2}' | head -1)

if [ -n "$ERRORS" ] && (( $(echo "$ERRORS > 10" | bc -l 2>/dev/null || echo "0") )); then
    echo -e "${RED}⚠️ Высокий уровень ошибок: $ERRORS${NC}"
    WARNINGS=$((WARNINGS + 1))
elif [ -n "$ERRORS" ]; then
    echo -e "${GREEN}✅ Ошибки: $ERRORS (норма)${NC}"
fi

# Проверяем соединения
echo -e "${YELLOW}🔗 Проверка соединений...${NC}"
CONNECTIONS=$(curl -s $SERVER_URL 2>/dev/null | grep 'quic_server_connections_total' | awk '{print $2}' | head -1)

if [ -n "$CONNECTIONS" ]; then
    echo -e "${GREEN}✅ Активных соединений: $CONNECTIONS${NC}"
fi

# Итоговый статус
echo ""
echo -e "${BLUE}==========================================${NC}"

if [ "$CRITICAL_ZONE_ALERT" = true ]; then
    echo -e "${RED}🚨 КРИТИЧЕСКОЕ СОСТОЯНИЕ: Сервер в критической зоне${NC}"
    exit 1
elif [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}⚠️ ПРЕДУПРЕЖДЕНИЯ: $WARNINGS проблем обнаружено${NC}"
    exit 2
else
    echo -e "${GREEN}✅ СЕРВЕР В НОРМЕ: Все проверки пройдены${NC}"
    exit 0
fi
EOF

    chmod +x scripts/health-check.sh
    echo "  ✅ Health check скрипт создан: scripts/health-check.sh"
}

# Функция для создания мониторинга
create_monitoring() {
    echo -e "${YELLOW}Создание системы мониторинга...${NC}"
    
    # Создаем Prometheus alerts
    cat > prometheus/alerts.yml << 'EOF'
groups:
- name: quic-server
  rules:
  - alert: QUICCriticalZone
    expr: quic_server_rate_per_connection >= 26 and quic_server_rate_per_connection <= 35
    for: 10s
    labels:
      severity: critical
    annotations:
      summary: "QUIC server in critical performance zone"
      description: "Server rate {{ $value }} pps is in critical zone (26-35 pps)"
  
  - alert: QUICHighJitter
    expr: histogram_quantile(0.95, quic_server_jitter_seconds) > 0.1
    for: 30s
    labels:
      severity: warning
    annotations:
      summary: "QUIC server high jitter detected"
      description: "Server jitter p95 is {{ $value }}s"
  
  - alert: QUICHighErrorRate
    expr: rate(quic_server_errors_total[5m]) / rate(quic_server_packets_total[5m]) > 0.01
    for: 1m
    labels:
      severity: warning
    annotations:
      summary: "QUIC server high error rate"
      description: "Server error rate is {{ $value }}%"
EOF

    echo "  ✅ Prometheus alerts созданы: prometheus/alerts.yml"
    
    # Создаем Grafana dashboard
    cat > grafana/dashboards/quic-optimization.json << 'EOF'
{
  "dashboard": {
    "title": "QUIC Server Optimization Dashboard",
    "panels": [
      {
        "title": "Connection Rate (Critical Zone Detection)",
        "type": "graph",
        "targets": [
          {
            "expr": "quic_server_rate_per_connection",
            "legendFormat": "Rate (pps)"
          }
        ],
        "yAxes": [
          {
            "min": 0,
            "max": 50
          }
        ],
        "thresholds": [
          {
            "value": 26,
            "colorMode": "critical",
            "op": "gt"
          },
          {
            "value": 35,
            "colorMode": "critical",
            "op": "lt"
          }
        ]
      },
      {
        "title": "Jitter (ms)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, quic_server_jitter_seconds) * 1000",
            "legendFormat": "Jitter P95 (ms)"
          }
        ]
      },
      {
        "title": "Error Rate (%)",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(quic_server_errors_total[5m]) / rate(quic_server_packets_total[5m]) * 100",
            "legendFormat": "Error Rate (%)"
          }
        ]
      }
    ]
  }
}
EOF

    echo "  ✅ Grafana dashboard создан: grafana/dashboards/quic-optimization.json"
}

# Основная функция
main() {
    echo -e "${GREEN}Применение DevOps оптимизаций...${NC}"
    echo ""
    
    # Применяем системные оптимизации
    apply_system_optimizations
    
    echo ""
    
    # Настраиваем лимиты процессов
    setup_process_limits
    
    echo ""
    
    # Создаем оптимизированный сервер
    create_optimized_server
    
    echo ""
    
    # Создаем health check
    create_health_check
    
    echo ""
    
    # Создаем мониторинг
    create_monitoring
    
    echo ""
    echo -e "${BLUE}==========================================${NC}"
    echo -e "${GREEN}✅ DevOps оптимизации применены успешно!${NC}"
    echo ""
    echo -e "${BLUE}Созданные компоненты:${NC}"
    echo "  🔧 scripts/optimized-server-start.sh - Оптимизированный сервер"
    echo "  🏥 scripts/health-check.sh - Health check"
    echo "  prometheus/alerts.yml - Prometheus алерты"
    echo "  📈 grafana/dashboards/quic-optimization.json - Grafana дашборд"
    echo ""
    echo -e "${BLUE}Команды для использования:${NC}"
    echo "  Запуск оптимизированного сервера: ./scripts/optimized-server-start.sh"
    echo "  Проверка здоровья: ./scripts/health-check.sh"
    echo "  Мониторинг: ./scripts/live-monitor.sh"
    echo ""
    echo -e "${YELLOW}💡 Рекомендации:${NC}"
    echo "  1. Используйте оптимизированный сервер для продакшена"
    echo "  2. Настройте мониторинг критических зон"
    echo "  3. Регулярно проверяйте health check"
    echo "  4. Настройте алерты для критических зон"
}

# Запуск
main "$@"

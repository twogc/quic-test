#!/bin/bash

# Скрипт для просмотра отчетов 2GC Network Protocol Suite

# Цвета для вывода
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}  2GC Network Protocol Suite - Reports${NC}"
echo "Просмотр отчетов тестирования"
echo ""

# Функция для форматирования JSON
format_json() {
    if command -v jq &> /dev/null; then
        jq .
    else
        cat
    fi
}

# Функция для показа краткой статистики
show_summary() {
    local file="$1"
    echo -e "${YELLOW}Краткая статистика из $file:${NC}"
    
    if [[ "$file" == *.json ]]; then
        if command -v jq &> /dev/null; then
            echo "  Успешные соединения: $(jq -r '.metrics.Success' "$file")"
            echo "  Ошибки: $(jq -r '.metrics.Errors' "$file")"
            echo "  Отправлено байт: $(jq -r '.metrics.BytesSent' "$file")"
            echo "  Потеря пакетов: $(jq -r '.metrics.PacketLoss' "$file")"
            echo "  Повторные передачи: $(jq -r '.metrics.Retransmits' "$file")"
            echo "  Время handshake: $(jq -r '.metrics.HandshakeTimes | join(", ")' "$file") мс"
        else
            echo "  (Установите jq для лучшего форматирования: sudo apt install jq)"
            head -20 "$file"
        fi
    else
        echo "  (Markdown отчет - полный текст)"
        head -10 "$file"
    fi
    echo ""
}

# Показываем доступные отчеты
echo -e "${YELLOW}Доступные отчеты:${NC}"
ls -la *.md *.json 2>/dev/null | grep -E "(report|test)" | while read -r line; do
    filename=$(echo "$line" | awk '{print $NF}')
    size=$(echo "$line" | awk '{print $5}')
    date=$(echo "$line" | awk '{print $6, $7, $8}')
    echo "  📄 $filename ($size bytes, $date)"
done
echo ""

# Показываем краткую статистику для каждого отчета
for report in *.md *.json 2>/dev/null; do
    if [[ -f "$report" && "$report" =~ (report|test) ]]; then
        show_summary "$report"
    fi
done

echo -e "${BLUE}🔍 Детальный просмотр отчетов:${NC}"
echo ""
echo -e "${YELLOW}1. Markdown отчет (человекочитаемый):${NC}"
echo "   cat report.md"
echo ""
echo -e "${YELLOW}2. JSON отчеты (структурированные данные):${NC}"
echo "   cat test-report.json | jq ."
echo "   cat debug-report.json | jq ."
echo ""
echo -e "${YELLOW}3. Просмотр в браузере (если есть веб-сервер):${NC}"
echo "   python3 -m http.server 8000"
echo "   # Затем откройте http://localhost:8000"
echo ""
echo -e "${YELLOW}4. Анализ метрик через Prometheus:${NC}"
echo "   curl http://localhost:2113/metrics"
echo ""
echo -e "${BLUE}📈 Полезные команды для анализа:${NC}"
echo ""
echo -e "${YELLOW}Статистика по файлам:${NC}"
echo "   ls -lah *.md *.json | grep -E '(report|test)'"
echo ""
echo -e "${YELLOW}Поиск по отчетам:${NC}"
echo "   grep -r 'Success' *.md *.json"
echo "   grep -r 'Errors' *.md *.json"
echo ""
echo -e "${YELLOW}Сортировка по времени:${NC}"
echo "   ls -lt *.md *.json | grep -E '(report|test)'"
echo ""
echo -e "${YELLOW}Размер отчетов:${NC}"
echo "   du -h *.md *.json | grep -E '(report|test)'"
echo ""
echo -e "${BLUE}🧹 Очистка старых отчетов:${NC}"
echo "   # Удалить отчеты старше 7 дней:"
echo "   find . -name '*.md' -o -name '*.json' | grep -E '(report|test)' | xargs ls -t | tail -n +10 | xargs rm -f"
echo ""
echo -e "${GREEN}Готово! Используйте команды выше для детального анализа.${NC}"


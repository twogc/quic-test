// 2GC CloudBridge QUICK testing - Dashboard JavaScript

class QuickTestDashboard {
    constructor() {
        this.isRunning = false;
        this.metrics = {};
        this.charts = {};
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadMetrics();
        this.setupCharts();
    }

    setupEventListeners() {
        // Кнопки управления
        document.getElementById('startTest')?.addEventListener('click', () => this.startTest());
        document.getElementById('stopTest')?.addEventListener('click', () => this.stopTest());
        document.getElementById('exportReport')?.addEventListener('click', () => this.exportReport());

        // Настройки теста
        document.getElementById('connections')?.addEventListener('change', (e) => {
            this.updateTestConfig('connections', parseInt(e.target.value));
        });

        document.getElementById('streams')?.addEventListener('change', (e) => {
            this.updateTestConfig('streams', parseInt(e.target.value));
        });

        document.getElementById('rate')?.addEventListener('change', (e) => {
            this.updateTestConfig('rate', parseInt(e.target.value));
        });

        document.getElementById('duration')?.addEventListener('change', (e) => {
            this.updateTestConfig('duration', parseInt(e.target.value));
        });
    }

    async loadMetrics() {
        try {
            const response = await fetch('/api/status');
            const data = await response.json();
            this.updateMetrics(data);
        } catch (error) {
            console.error('Ошибка загрузки метрик:', error);
        }
    }

    updateMetrics(data) {
        // Обновление статуса
        const statusElement = document.getElementById('status');
        if (statusElement) {
            statusElement.innerHTML = `
                <span class="status-indicator ${data.server?.running ? 'status-running' : 'status-stopped'}"></span>
                Сервер: ${data.server?.running ? 'Запущен' : 'Остановлен'}
            `;
        }

        // Обновление метрик
        if (data.metrics) {
            this.updateMetricCard('connections', data.metrics.connections || 0);
            this.updateMetricCard('streams', data.metrics.streams || 0);
            this.updateMetricCard('throughput', data.metrics.throughput || 0);
            this.updateMetricCard('latency', data.metrics.latency || 0);
            this.updateMetricCard('errors', data.metrics.errors || 0);
        }
    }

    updateMetricCard(id, value) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = value;
        }
    }

    setupCharts() {
        // Инициализация графиков (заглушка)
        // Здесь можно интегрировать Chart.js или другую библиотеку
        console.log('Инициализация графиков...');
    }

    async startTest() {
        if (this.isRunning) return;

        try {
            const response = await fetch('/api/run', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(this.getTestConfig())
            });

            if (response.ok) {
                this.isRunning = true;
                this.updateUI();
                this.startMetricsPolling();
            } else {
                throw new Error('Ошибка запуска теста');
            }
        } catch (error) {
            console.error('Ошибка запуска теста:', error);
            alert('Ошибка запуска теста: ' + error.message);
        }
    }

    async stopTest() {
        if (!this.isRunning) return;

        try {
            const response = await fetch('/api/stop', {
                method: 'POST'
            });

            if (response.ok) {
                this.isRunning = false;
                this.updateUI();
                this.stopMetricsPolling();
            }
        } catch (error) {
            console.error('Ошибка остановки теста:', error);
        }
    }

    getTestConfig() {
        return {
            connections: parseInt(document.getElementById('connections')?.value || 1),
            streams: parseInt(document.getElementById('streams')?.value || 1),
            rate: parseInt(document.getElementById('rate')?.value || 100),
            duration: parseInt(document.getElementById('duration')?.value || 0),
            packetSize: parseInt(document.getElementById('packetSize')?.value || 1200),
            pattern: document.getElementById('pattern')?.value || 'random'
        };
    }

    updateTestConfig(key, value) {
        // Обновление конфигурации теста
        console.log(`Обновление конфигурации: ${key} = ${value}`);
    }

    updateUI() {
        const startBtn = document.getElementById('startTest');
        const stopBtn = document.getElementById('stopTest');

        if (this.isRunning) {
            startBtn?.classList.add('disabled');
            stopBtn?.classList.remove('disabled');
        } else {
            startBtn?.classList.remove('disabled');
            stopBtn?.classList.add('disabled');
        }
    }

    startMetricsPolling() {
        this.metricsInterval = setInterval(() => {
            this.loadMetrics();
        }, 1000);
    }

    stopMetricsPolling() {
        if (this.metricsInterval) {
            clearInterval(this.metricsInterval);
            this.metricsInterval = null;
        }
    }

    async exportReport() {
        try {
            const response = await fetch('/api/report');
            const blob = await response.blob();
            
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `quck-test-report-${new Date().toISOString().slice(0, 19)}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        } catch (error) {
            console.error('Ошибка экспорта отчета:', error);
            alert('Ошибка экспорта отчета: ' + error.message);
        }
    }
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    new QuickTestDashboard();
});

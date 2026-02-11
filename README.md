# ADC - API Declarative CLI for APISIX

ADC (API Declarative CLI) - это инструмент командной строки для декларативного управления конфигурацией Apache APISIX через YAML/JSON файлы.

## ✨ Возможности

- 🚀 Декларативное управление конфигурацией APISIX
- 📝 Поддержка YAML и JSON форматов
- 🔄 Синхронизация локальной и удаленной конфигурации
- 💾 Автоматические backups перед изменениями
- 🔍 Diff для просмотра изменений
- ✅ Валидация конфигурации
- 🔁 Retry логика с exponential backoff
- 📦 Поддержка всех ресурсов APISIX:
  - Routes
  - Services
  - Upstreams
  - Consumers
  - SSLs
  - Global Rules
  - Plugin Configs
  - Stream Routes
  - Plugin Metadata

## 📦 Установка

### Из исходников

```bash
git clone https://github.com/api7/adc-go.git
cd adc-go
go build -o bin/adc ./cmd/adc
sudo mv bin/adc /usr/local/bin/
```

### Проверка установки

```bash
adc version
```

## 🚀 Быстрый старт

### 1. Инициализация

```bash
adc init
```

Эта команда создаст:
- Конфигурационный файл: `~/.config/adc/config.yaml`
- Пример декларативной конфигурации: `example-config.yaml`

### 2. Настройка подключения к APISIX

Отредактируйте `~/.config/adc/config.yaml`:

```yaml
version: "1.0.0"
debug: false
apisix:
  base_url: "http://127.0.0.1:9180"
  admin_key: "ваш-admin-key"
  admin_key_name: "X-API-Key"
```

Или используйте переменные окружения:

```bash
export ADC_APISIX_BASE_URL="http://127.0.0.1:9180"
export ADC_APISIX_ADMIN_KEY="ваш-admin-key"
```

### 3. Проверка подключения

```bash
adc ping
```

### 4. Создание конфигурации

Создайте файл `my-config.yaml`:

```yaml
version: "1.0"

routes:
  - id: "api-route"
    name: "API Route"
    uri: "/api/*"
    methods: ["GET", "POST"]
    upstream:
      id: "api-upstream"
      nodes:
        "backend:8080": 1
    plugins:
      cors: {}
      rate-limit:
        count: 100
        time_window: 60
```

### 5. Валидация

```bash
adc validate -f my-config.yaml
```

### 6. Просмотр изменений

```bash
adc diff -f my-config.yaml
```

### 7. Применение конфигурации

```bash
# Dry-run (без применения)
adc apply -f my-config.yaml --dry-run

# Реальное применение
adc apply -f my-config.yaml
```

## 📚 Команды

### Основные команды

| Команда | Описание |
|---------|----------|
| `adc init` | Инициализация конфигурации |
| `adc version` | Вывод версии |
| `adc ping` | Проверка подключения к APISIX |
| `adc validate -f FILE` | Валидация конфигурации |
| `adc diff -f FILE` | Показать различия |
| `adc apply -f FILE` | Применить конфигурацию |
| `adc sync -f FILE` | Синхронизация (с удалением) |
| `adc dump` | Экспорт текущей конфигурации |

### Backup команды

| Команда | Описание |
|---------|----------|
| `adc backup` | Создать backup |
| `adc backup-list` | Список backups |
| `adc restore [FILE]` | Восстановить из backup |
| `adc backup-delete FILE` | Удалить backup |

## 🔧 Примеры использования

### Экспорт текущей конфигурации

```bash
# В YAML
adc dump -o current.yaml

# В JSON
adc dump -o current.json --format json

# В stdout
adc dump
```

### Применение с опциями

```bash
# Без подтверждения
adc apply -f config.yaml --force

# Без автоматического backup
adc apply -f config.yaml --no-backup

# Dry-run
adc apply -f config.yaml --dry-run
```

### Работа с backups

```bash
# Создать backup с описанием
adc backup --description "Before major update"

# Список всех backups
adc backup-list

# Восстановить конкретный backup
adc restore backup-20260211-143000.yaml

# Восстановить без подтверждения
adc restore backup-20260211-143000.yaml --force
```

### Синхронизация

```bash
# Полная синхронизация (удалит ресурсы, которых нет в файле)
adc sync -f config.yaml

# Без подтверждения
adc sync -f config.yaml --force
```

## 📖 Структура конфигурации

### Полный пример

```yaml
version: "1.0"

# Upstreams (апстримы)
upstreams:
  - id: "upstream-1"
    name: "Backend Service"
    type: "roundrobin"
    nodes:
      "backend-1:8080": 1
      "backend-2:8080": 1
    checks:
      active:
        type: "http"
        http_path: "/health"
        healthy:
          interval: 2
          successes: 2

# Services (сервисы)
services:
  - id: "service-1"
    name: "API Service"
    upstream_id: "upstream-1"
    plugins:
      rate-limit:
        count: 100
        time_window: 60

# Routes (маршруты)
routes:
  - id: "route-1"
    name: "API Route"
    uri: "/api/v1/*"
    methods: ["GET", "POST"]
    service_id: "service-1"
    plugins:
      cors: {}

# Consumers (потребители)
consumers:
  - username: "user-1"
    plugins:
      key-auth:
        key: "secret-key-123"

# SSL Certificates
ssls:
  - id: "ssl-1"
    cert: "-----BEGIN CERTIFICATE-----\n...\n"
    key: "-----BEGIN PRIVATE KEY-----\n...\n"
    snis:
      - "example.com"

# Global Rules
global_rules:
  - id: "global-1"
    plugins:
      prometheus: {}

# Plugin Configs
plugin_configs:
  - id: "pc-1"
    desc: "Common plugins"
    plugins:
      cors: {}
      gzip: {}

# Stream Routes
stream_routes:
  - id: "stream-1"
    server_port: 9100
    upstream:
      id: "stream-upstream-1"
      nodes:
        "tcp-backend:3306": 1
```

## 🔐 Безопасность

### Хранение credentials

Рекомендуется использовать переменные окружения для хранения sensitive данных:

```bash
export ADC_APISIX_ADMIN_KEY="ваш-секретный-ключ"
```

### Backup безопасность

Backups содержат полную конфигурацию, включая SSL ключи и credentials. Храните их в безопасном месте:

```bash
# Изменить директорию backups
adc backup --backup-dir /secure/path/backups
```

## 🛠️ Разработка

### Структура проекта

```
adc-go/
├── cmd/adc/           # Точка входа
├── internal/
│   ├── apisix/        # APISIX HTTP client
│   ├── backup/        # Backup & restore
│   ├── commands/      # CLI команды
│   ├── config/        # Конфигурация
│   ├── declarative/   # Типы данных
│   ├── diff/          # Diff логика
│   ├── retry/         # Retry механизм
│   ├── sync/          # Sync логика
│   └── utils/         # Утилиты
├── go.mod
└── go.sum
```

### Сборка

```bash
go build -o bin/adc ./cmd/adc
```

### Тестирование

```bash
go test ./...
```

## 📝 Конфигурация

### ADC конфигурация

Файл: `~/.config/adc/config.yaml`

```yaml
version: "1.0.0"
debug: false
apisix:
  base_url: "http://127.0.0.1:9180"
  admin_key: "edd1c9f034335f136f87ad84b625c8f1"
  admin_key_name: "X-API-Key"
```

### Переменные окружения

| Переменная | Описание |
|------------|----------|
| `ADC_APISIX_BASE_URL` | URL APISIX Admin API |
| `ADC_APISIX_ADMIN_KEY` | Admin API ключ |
| `ADC_APISIX_ADMIN_KEY_NAME` | Имя header для ключа |
| `ADC_DEBUG` | Режим отладки (true/false) |

## 🤝 Вклад в проект

Мы приветствуем вклад в проект! Пожалуйста:

1. Fork репозиторий
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📄 Лицензия

Apache License 2.0

## 🙏 Благодарности

- [Apache APISIX](https://apisix.apache.org/)
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management

## 📞 Поддержка

- GitHub Issues: [https://github.com/api7/adc-go/issues](https://github.com/api7/adc-go/issues)
- APISIX Slack: [https://apisix.apache.org/docs/general/join/](https://apisix.apache.org/docs/general/join/)

## 🗺️ Roadmap

- [ ] Unit и интеграционные тесты
- [ ] JSON Schema валидация
- [ ] Template поддержка
- [ ] Watch mode
- [ ] Цветной вывод
- [ ] Progress bar
- [ ] Автодополнение для shell
- [ ] Docker образ
- [ ] Helm chart

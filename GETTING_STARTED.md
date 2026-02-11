# Getting Started with ADC

Это руководство поможет вам начать работу с ADC (API Declarative CLI) для управления Apache APISIX.

## Предварительные требования

- Go 1.21 или выше (для сборки из исходников)
- Запущенный Apache APISIX с доступным Admin API
- Admin API ключ для APISIX

## Установка

### Вариант 1: Сборка из исходников

```bash
# Клонировать репозиторий
git clone https://github.com/api7/adc-go.git
cd adc-go

# Собрать
make build

# Установить
sudo make install
```

### Вариант 2: Использование go install

```bash
go install github.com/api7/adc-go/cmd/adc@latest
```

### Проверка установки

```bash
adc version
```

Вы должны увидеть:
```
ADC Version: dev
Build Time: 2026-02-11
Git Commit: unknown
```

## Первоначальная настройка

### Шаг 1: Инициализация

```bash
adc init
```

Эта команда создаст:
- `~/.config/adc/config.yaml` - конфигурационный файл ADC
- `example-config.yaml` - пример декларативной конфигурации

### Шаг 2: Настройка подключения к APISIX

Отредактируйте `~/.config/adc/config.yaml`:

```yaml
version: "1.0.0"
debug: false
apisix:
  base_url: "http://127.0.0.1:9180"  # URL вашего APISIX Admin API
  admin_key: "ваш-admin-key"          # Ваш Admin API ключ
  admin_key_name: "X-API-Key"         # Имя header (обычно не нужно менять)
```

**Важно:** Замените `ваш-admin-key` на реальный ключ из конфигурации APISIX.

#### Альтернатива: Использование переменных окружения

Вместо хранения ключа в файле, можно использовать переменные окружения:

```bash
export ADC_APISIX_BASE_URL="http://127.0.0.1:9180"
export ADC_APISIX_ADMIN_KEY="ваш-admin-key"
```

### Шаг 3: Проверка подключения

```bash
adc ping
```

Если все настроено правильно, вы увидите:
```
Connecting to APISIX at: http://127.0.0.1:9180
✓ Successfully connected to APISIX Admin API
```

## Первая конфигурация

### Создание простого маршрута

Создайте файл `my-first-route.yaml`:

```yaml
version: "1.0"

routes:
  - id: "hello-world"
    name: "Hello World Route"
    uri: "/hello"
    methods: ["GET"]
    upstream:
      id: "httpbin"
      nodes:
        "httpbin.org:80": 1
```

### Валидация конфигурации

```bash
adc validate -f my-first-route.yaml
```

Вы должны увидеть:
```
✓ Configuration file my-first-route.yaml is valid
  Version: 1.0
  Routes: 1
  Services: 0
  Upstreams: 1
  Consumers: 0
  SSLs: 0

  Routes details:
    - hello-world: Hello World Route (upstream: httpbin)
```

### Просмотр изменений

Перед применением конфигурации, посмотрите что изменится:

```bash
adc diff -f my-first-route.yaml
```

Вы увидите список изменений:
```
Comparing with APISIX at: http://127.0.0.1:9180

Fetching current APISIX state...
Changes to be applied:

Routes:
  + Create (1):
    + hello-world

Upstreams:
  + Create (1):
    + httpbin
```

### Применение конфигурации (Dry-run)

Сначала попробуйте dry-run режим:

```bash
adc apply -f my-first-route.yaml --dry-run
```

Это покажет что будет сделано, но не применит изменения.

### Применение конфигурации

Теперь примените конфигурацию:

```bash
adc apply -f my-first-route.yaml
```

ADC:
1. Создаст автоматический backup текущей конфигурации
2. Покажет изменения
3. Попросит подтверждение
4. Применит изменения

Вы увидите:
```
Applying configuration to: http://127.0.0.1:9180
Fetching current APISIX state...
Calculating differences...
Changes to be applied:

Routes:
  + Create (1):
    + hello-world

Creating backup before applying changes...
✓ Backup created: /home/user/.config/adc/backups/backup-20260211-143000.yaml

Do you want to apply these changes? (yes/no): yes

Applying changes...
✓ Created upstream: httpbin
✓ Created route: hello-world

✓ Configuration applied successfully!
```

### Проверка маршрута

Теперь ваш маршрут доступен через APISIX:

```bash
curl http://localhost:9080/hello
```

## Основные операции

### Экспорт текущей конфигурации

```bash
# В YAML
adc dump -o current-config.yaml

# В JSON
adc dump -o current-config.json --format json
```

### Обновление конфигурации

Отредактируйте `my-first-route.yaml`, например, добавьте плагин:

```yaml
version: "1.0"

routes:
  - id: "hello-world"
    name: "Hello World Route"
    uri: "/hello"
    methods: ["GET"]
    upstream:
      id: "httpbin"
      nodes:
        "httpbin.org:80": 1
    plugins:
      cors: {}  # Добавили CORS плагин
```

Примените изменения:

```bash
adc apply -f my-first-route.yaml
```

ADC автоматически определит что нужно обновить маршрут.

### Работа с backups

#### Список backups

```bash
adc backup-list
```

#### Создание backup вручную

```bash
adc backup --description "Before major changes"
```

#### Восстановление из backup

```bash
# Показать доступные backups
adc restore

# Восстановить конкретный backup
adc restore backup-20260211-143000.yaml
```

### Синхронизация (с удалением)

Команда `sync` удалит все ресурсы, которых нет в вашем файле:

```bash
adc sync -f my-first-route.yaml
```

**Внимание:** Эта команда удалит все маршруты, сервисы и другие ресурсы, которые не указаны в файле!

## Продвинутые примеры

### Пример с сервисом и несколькими маршрутами

Создайте `api-config.yaml`:

```yaml
version: "1.0"

upstreams:
  - id: "backend-api"
    name: "Backend API"
    type: "roundrobin"
    nodes:
      "backend-1:8080": 1
      "backend-2:8080": 1
    checks:
      active:
        http_path: "/health"
        healthy:
          interval: 2
          successes: 2

services:
  - id: "api-service"
    name: "API Service"
    upstream_id: "backend-api"
    plugins:
      rate-limit:
        count: 100
        time_window: 60

routes:
  - id: "api-v1"
    name: "API v1"
    uri: "/api/v1/*"
    methods: ["GET", "POST", "PUT", "DELETE"]
    service_id: "api-service"
    plugins:
      cors: {}

  - id: "api-v2"
    name: "API v2"
    uri: "/api/v2/*"
    methods: ["GET", "POST"]
    service_id: "api-service"
```

Примените:

```bash
adc apply -f api-config.yaml
```

### Пример с аутентификацией

Создайте `auth-config.yaml`:

```yaml
version: "1.0"

consumers:
  - username: "api-user"
    plugins:
      key-auth:
        key: "secret-api-key-123"

routes:
  - id: "protected-api"
    name: "Protected API"
    uri: "/protected/*"
    methods: ["GET", "POST"]
    upstream:
      id: "backend"
      nodes:
        "backend:8080": 1
    plugins:
      key-auth: {}
```

Примените:

```bash
adc apply -f auth-config.yaml
```

Теперь для доступа к `/protected/*` нужен API ключ:

```bash
curl -H "apikey: secret-api-key-123" http://localhost:9080/protected/resource
```

## Полезные команды

### Применение без подтверждения

```bash
adc apply -f config.yaml --force
```

### Применение без автоматического backup

```bash
adc apply -f config.yaml --no-backup
```

### Просмотр конфигурации в stdout

```bash
adc dump
```

### Использование другого конфигурационного файла

```bash
adc --config /path/to/config.yaml ping
```

### Режим отладки

```bash
adc --debug apply -f config.yaml
```

## Troubleshooting

### Ошибка подключения

```
Error: failed to connect to APISIX: connection refused
```

**Решение:**
- Проверьте что APISIX запущен
- Проверьте URL в конфигурации
- Проверьте что Admin API доступен

### Ошибка аутентификации

```
Error: API request failed with status 401: Unauthorized
```

**Решение:**
- Проверьте admin_key в конфигурации
- Проверьте что ключ совпадает с конфигурацией APISIX

### Ошибка валидации

```
Error: route at index 0: uri or uris is required
```

**Решение:**
- Проверьте что все обязательные поля заполнены
- Используйте `adc validate` для проверки конфигурации

## Следующие шаги

1. Изучите [полные примеры](examples/) конфигураций
2. Прочитайте [документацию APISIX](https://apisix.apache.org/docs/)
3. Настройте CI/CD для автоматического применения конфигурации
4. Используйте Git для версионирования конфигурации

## Полезные ссылки

- [README.md](README.md) - Основная документация
- [FINAL_STATUS.md](FINAL_STATUS.md) - Полный список функций
- [examples/](examples/) - Примеры конфигураций
- [APISIX Documentation](https://apisix.apache.org/docs/)

## Получение помощи

Если у вас возникли проблемы:

1. Проверьте документацию
2. Используйте `--debug` флаг для детальной информации
3. Создайте issue на GitHub
4. Спросите в APISIX Slack

Удачи в использовании ADC! 🚀

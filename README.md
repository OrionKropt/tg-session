# Telegram Session Service
Микросервис на Go для управления множественными независимыми сессиями Telegram через протокол MTProto. Сервис позволяет динамически создавать соединения, авторизоваться через QR-код и обмениваться сообщениями через gRPC интерфейс.

## Функционал
- поддержка неограниченного количества изолированных подключений
- QR-авторизация
- отправка текстовых сообщений
- получение входящих сообщений
- гарантия доставки пропущенных обновлений
- изоляция сессий
- graceful shutdown - завершение всех сессий c выходом из Telegram при остановке сервиса

## Архитектура
Сервис спроектирован по принципам *Clean Architecture*:

- Domain: Бизнес-модели и интерфейсы
- UseCase (Core): Оркестрация логики, управление жизненным циклом сессий
- Infrastructure: Реализация Telegram клиента (gotd) и репозитория
- Controller: gRPC сервер с поддержкой Server Streaming для получения сообщений

## Стек
- Go 1.26
- github.com/gotd/td
- github.com/google/uuid
- gRPC
- Protocol Buffers
- Docker / Docker Compose

## Конфигурация
### Создайте файл .env в корне проекта на основе примера:
```env
TELEGRAM_APP_ID=12345
TELEGRAM_APP_HASH=your_hash_here
TELEGRAM_PEER_DATABASE_NAME=peers.db
LOG_LEVEL=info
```
Получить `APP_ID` и `APP_HASH` можно на https://my.telegram.org/

## Запуск сервиса
### 1. Запуск через Docker (Рекомендуется)
```bash
make dev-docker-compose-up
```
Остановить:
```bash
make docker-compose-down
```
Сервис будет доступен на:
```
localhost:50051
```

### 2.Локальный запуск
```
make tidy
export $(cat .env | xargs) && go run cmd/app/main.go
```

## gRPC API
### CreateSession
Создание новой Telegram-сессии.

Возвращает:
- `session_id`
- `qr_code` (URL для генерации QR-кода)

Пример запроса
```json
{}
```

Пример через grpcurl:
```bash
grpcurl -plaintext -d '{}' localhost:50051 telegram.v1.TGSession/CreateSession
```

Ответ:
```json
{
  "sessionId": "uuid",
  "qrCode": "tg://login?token=..."
}
```

### SendMessage
Отправка текстового сообщения.

Пример запроса:
```json
{
  "session_id": "uuid",
  "peer": "username_or_channel",
  "text": "Hello from gRPC"
}
```

Пример через grpcsurl:
```bash
grpcurl -plaintext -d '{
  "session_id": "uuid",
  "peer": "username_or_channel",
  "text": "Hello from gRPC"
}' localhost:50051 telegram.v1.TGSession/SendMessage
```

Ответ:
```json
{
  "messageId": 123456
}
```

### SubscribeMessages (Server Streaming)
Подписка на входящие сообщения конкретной сессии.

Пример запроса
```json
{
  "session_id": "uuid"
}
```

Пример через grpcsurl:
```bash
grpcurl -plaintext -d '{
  "session_id": "uuid"
}' localhost:50051 telegram.v1.TGSession/SubscribeMessages
```

Стрим возвращает:
```json
{
  "from": "@username",
  "messageId": "283500",
  "text": "text",
  "timestamp": "1772597010"
}
```

### DeleteSession
Удаляет соединение:

- Останавливает Telegram клиент
- Выполняет logout (если авторизован)
- Освобождает ресурсы

Пример запроса:
```json
{
  "session_id": "uuid"
}
```

Пример через grpcsurl:
```bash
grpcurl -plaintext -d '{
  "session_id": "uuid"
}' localhost:50051 telegram.v1.TGSession/DeleteSession
```

## Изоляция соединений
Каждая сессия имеет собственный Telegram-клиент c update dispatcher и каналом сообщений. У всех сессий собсвтенный жизненный цикл - ошибка или отключение одной сессии не влияет на остальные.

## Graceful Shutdown
При остановке сервиса все активные сессии корректно закрываются, выполянется процесс logout, останавливатеся gRPC сервер.

## Процесс авторизации через QR
Авторизация происходит в несколько этапов:
1. Клиент вызывает `CreateSession`
2. Сервис:
    - cоздаёт новый Telegram-клиент
    - устанавливает соединение с MTProto
    - запускает QR-авторизацию
    - возвращает `qrCode` (URL-токен)

3. Клиент генерирует QR-код из полученного URL
4. Пользователь:
    - открывает Telegram
    - переходит в Settings → Devices → Scan QR
    - сканирует QR-код
5. После успешного сканирования Telegram завершает авторизацию
6. Сессия становится активной и готовой к отправке/получению сообщений

### Схема взаимодействия
```
gRPC Client
    │
    │ 1. CreateSession
    ▼
Telegram Session Service
    │
    │ 2. Connect (gotd)
    │ 3. Start QR auth
    ▼
Telegram MTProto
    ▲
    │ 4. Scan QR (Mobile App)
    │
User (Telegram App)
```
После авторизации:
```
gRPC Client ── SendMessage ──▶ Service ──▶ Telegram
Telegram ── Updates ──▶ Service ──▶ SubscribeMessages (stream) ──▶ gRPC Client
```
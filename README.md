# PC Club Server

Мини REST API на Go для WPF-приложения администрирования компьютерного клуба.

## Быстрый старт

```powershell
go mod tidy
go run ./cmd/server
```

По умолчанию сервер слушает `http://localhost:8080`, база SQLite создается в `data/pcclub.db`.

## Railway

1. Создай новый GitHub-репозиторий из этой папки.
2. Подключи репозиторий в Railway.
3. Railway сам передаст переменную `PORT`.
4. Для постоянного хранения SQLite добавь volume и укажи:

```text
DATABASE_PATH=/app/data/pcclub.db
```

Полезные переменные окружения:

```text
PORT=8080
DATABASE_PATH=data/pcclub.db
SEED_DEMO_DATA=true
```

Если демо-данные не нужны:

```text
SEED_DEMO_DATA=false
```

## Эндпоинты

### Служебные

```http
GET /health
```

### Клиенты

```http
GET /api/clients
POST /api/clients
PUT /api/clients/{id}
DELETE /api/clients/{id}
```

Тело для `POST` и `PUT`:

```json
{
  "name": "Иванов Иван",
  "phone": "+7 (999) 123-45-67",
  "email": "ivan@mail.ru",
  "balance": 1500
}
```

### Компьютеры

```http
GET /api/computers
POST /api/computers
PUT /api/computers/{id}
DELETE /api/computers/{id}
```

```json
{
  "name": "ПК-01",
  "isOccupied": false
}
```

### Тарифы

```http
GET /api/tariffs
POST /api/tariffs
PUT /api/tariffs/{id}
DELETE /api/tariffs/{id}
```

```json
{
  "name": "Базовый",
  "costPerHour": 300,
  "description": "Стандартный тариф"
}
```

### Сессии

```http
GET /api/sessions
GET /api/sessions?computer=ПК-01
POST /api/sessions
POST /api/sessions/{id}/complete
DELETE /api/sessions/{id}
```

Тело для старта сессии:

```json
{
  "clientName": "Иванов Иван",
  "computerName": "ПК-01",
  "tariffName": "Базовый",
  "hourlyRate": 300
}
```

### Статистика

```http
GET /api/statistics
```

Возвращает общую выручку, количество завершенных и активных сессий, доход по дням, распределение по компьютерам и популярные тарифы.

## Локальная проверка

```powershell
go test ./...
go build ./cmd/server
```

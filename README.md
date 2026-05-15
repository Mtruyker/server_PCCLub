# PC Club Server

REST API на Go для компьютерного клуба. Сервер хранит данные в SQLite и подходит для мобильного клиента: авторизация, клиенты, карта ПК, бронирования, каталог, заказы, игровые сессии и новости.

## Быстрый Старт

```powershell
go mod tidy
go run ./cmd/server
```

По умолчанию сервер слушает `http://localhost:8080`, база SQLite создается в `data/pcclub.db`.

## Переменные Окружения

```text
PORT=8080
DATABASE_PATH=data/pcclub.db
SEED_DEMO_DATA=true
JWT_SECRET=change-me
ADMIN_API_TOKEN=admin-token
```

`JWT_SECRET` используется для подписи access/refresh токенов. Для production обязательно задай свое значение.

`ADMIN_API_TOKEN` - статический токен администратора для desktop/admin приложения. Для совместимости сервер также понимает `PCCLUB_API_TOKEN`; если обе переменные не заданы, он попробует прочитать токен из файла `api-token.txt` в рабочей папке.

Если демо-данные не нужны:

```text
SEED_DEMO_DATA=false
```

## Railway

1. Создай GitHub-репозиторий из этой папки.
2. Подключи репозиторий в Railway.
3. Railway сам передаст переменную `PORT`.
4. Для постоянного хранения SQLite добавь volume и укажи:

```text
DATABASE_PATH=/app/data/pcclub.db
JWT_SECRET=<strong-secret>
ADMIN_API_TOKEN=<admin-token>
```

## Авторизация

Защищенные клиентские методы ожидают заголовок:

```http
Authorization: Bearer <token>
```

Административное приложение может передавать статический токен так:

```http
Authorization: Bearer <admin-token>
```

или так:

```http
X-API-Token: <admin-token>
```

### Регистрация

```http
POST /api/auth/register
Content-Type: application/json
```

```json
{
  "name": "Anatoliy",
  "phone": "89962659984",
  "email": "",
  "password": "tester1"
}
```

Ответ:

```json
{
  "clientId": 5,
  "token": "jwt_access_token",
  "refreshToken": "jwt_refresh_token"
}
```

### Логин

```http
POST /api/auth/login
Content-Type: application/json
```

```json
{
  "phone": "89962659984",
  "password": "tester1"
}
```

Пароль хранится только как bcrypt hash и не возвращается в API.

## Клиенты

Публичные административные методы:

```http
GET /api/clients
POST /api/clients
DELETE /api/clients/{id}
```

Персональные методы требуют `Authorization`.

```http
GET /api/clients/{id}
PUT /api/clients/{id}
POST /api/clients/{id}/top-up
```

`GET /api/clients/{id}`:

```json
{
  "id": 5,
  "name": "Anatoliy",
  "phone": "89962659984",
  "email": "",
  "balance": 0
}
```

`PUT /api/clients/{id}`:

```json
{
  "name": "Anatoliy",
  "phone": "89962659984",
  "email": "test@mail.ru"
}
```

`POST /api/clients/{id}/top-up`:

```json
{
  "amount": 500
}
```

## ПК И Карта Клуба

Полный список ПК:

```http
GET /api/pcs
```

```json
[
  {
    "id": 1,
    "name": "ПК-01",
    "zone": "standard",
    "x": 50,
    "y": 150,
    "isOccupied": false,
    "status": "Свободен",
    "hourPrice": 120
  }
]
```

Список свободных ПК:

```http
GET /api/pcs/available
```

```json
[
  {
    "id": 1,
    "name": "ПК-01",
    "isOccupied": false,
    "status": "Свободен"
  }
]
```

Старые административные endpoints для компьютеров тоже сохранены:

```http
GET /api/computers
POST /api/computers
PUT /api/computers/{id}
DELETE /api/computers/{id}
```

## Бронирования

Все методы бронирований требуют `Authorization`.

```http
POST /api/bookings
GET /api/clients/{id}/bookings
DELETE /api/bookings/{id}
```

Создание брони:

```json
{
  "clientId": 5,
  "pcName": "ПК-01",
  "startTime": "2026-05-14T19:30:00.000Z",
  "duration": 2
}
```

Ответ:

```json
{
  "id": 12,
  "clientId": 5,
  "pcName": "ПК-01",
  "startTime": "2026-05-14T19:30:00.000Z",
  "durationHours": 2,
  "totalPrice": 240,
  "status": "active"
}
```

Сервер проверяет пересечения по времени. Если тот же ПК уже забронирован на пересекающийся интервал, вернется `409 Conflict`.

## Каталог

```http
GET /api/items
```

```json
[
  {
    "id": 1,
    "name": "Coca-Cola 0.5",
    "description": "Напиток",
    "price": 90,
    "imageUrl": "",
    "category": "Напитки"
  }
]
```

## Заказы

Методы заказов требуют `Authorization`.

```http
POST /api/orders
GET /api/clients/{id}/orders
```

Создание заказа:

```json
{
  "clientId": 5,
  "items": [
    {
      "productId": 1,
      "quantity": 2
    }
  ],
  "date": "2026-05-14T19:30:00.000Z"
}
```

Ответ:

```json
{
  "id": 501,
  "clientId": 5,
  "date": "2026-05-14T19:30:00.000Z",
  "items": [
    {
      "productName": "Coca-Cola 0.5",
      "quantity": 2,
      "price": 90
    }
  ],
  "totalAmount": 180,
  "status": "pending"
}
```

Поддерживаемые статусы в модели: `pending`, `preparing`, `delivering`, `delivered`, `cancelled`. Новые заказы создаются со статусом `pending`.

## Игровые Сессии

Клиентская история требует `Authorization`:

```http
GET /api/clients/{id}/sessions
```

```json
[
  {
    "id": 1,
    "pcName": "ПК-01",
    "startTime": "2026-05-14T18:00:00.000Z",
    "endTime": "2026-05-14T20:00:00.000Z",
    "cost": 240
  }
]
```

Старые административные endpoints для сессий сохранены:

```http
GET /api/sessions
GET /api/sessions?computer=ПК-01
POST /api/sessions
POST /api/sessions/{id}/complete
DELETE /api/sessions/{id}
```

## Новости

```http
GET /api/news
```

```json
[
  {
    "id": 1,
    "title": "Турнир в субботу",
    "content": "Описание новости",
    "imageUrl": "",
    "date": "2026-05-14T12:00:00.000Z"
  }
]
```

## Тарифы

Административные endpoints:

```http
GET /api/tariffs
POST /api/tariffs
PUT /api/tariffs/{id}
DELETE /api/tariffs/{id}
```

```json
{
  "name": "Базовый",
  "costPerHour": 120,
  "description": "Стандартное игровое место"
}
```

## Статистика

```http
GET /api/statistics
```

Возвращает общую выручку, количество завершенных и активных сессий, доход по дням, распределение по компьютерам и популярные тарифы.

## Минимальная Схема БД

```text
Clients(Id, Name, Phone, Email, PasswordHash, Balance)
Computers(Id, Name, Zone, X, Y, IsOccupied, HourPrice)
Items(Id, Name, Category, Price, ImageUrl, Description)
News(Id, Title, Content, ImageUrl, PublishedAt)
Bookings(Id, ClientId, ClientName, PcId, ComputerName, StartTime, EndTime, TariffName, HourlyRate, DurationHours, TotalPrice, Status)
Orders(Id, ClientId, Date, TotalAmount, Status)
OrderItems(Id, OrderId, ProductId, ProductName, Quantity, Price)
Sessions(Id, ClientId, ClientName, PcId, ComputerName, StartTime, EndTime, TotalCost, TariffName, HourlyRate)
Tariffs(Id, Name, CostPerHour, Description)
```

## Локальная Проверка

```powershell
go test ./...
go build ./cmd/server
```

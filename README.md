# PCClub Server

Оптимизированный REST API на Go для мобильного приложения компьютерного клуба. Сервер обеспечивает стабильную работу с высокой производительностью и полную совместимость с [PCCLUBMobile](https://github.com/Mtruyker/PCCLUBMobile).

## 🚀 Быстрый Старт

### Локальная разработка

```bash
# Настройка окружения
cp .env.example .env
# Отредактируйте .env с вашими настройками

# Запуск сервера
go mod tidy
go run ./cmd/server
```

### Docker развертывание

```bash
# Быстрое развертывание
chmod +x deploy.sh
./deploy.sh

# Или вручную
docker-compose up -d
```

По умолчанию сервер доступен на `http://localhost:8080`, база SQLite создается в `data/pcclub.db`.

## 📱 API v1 для мобильного приложения

### Аутентификация
- `POST /api/v1/auth/register` - Регистрация пользователя
- `POST /api/v1/auth/login` - Вход в систему  
- `POST /api/v1/auth/refresh` - Обновление JWT токена

### Профиль пользователя
- `GET /api/v1/profile` - Получить профиль текущего пользователя
- `PUT /api/v1/profile` - Обновить профиль

### Компьютеры и бронирование
- `GET /api/v1/pcs` - Список всех компьютеров с картой клуба
- `GET /api/v1/pcs/available` - Доступные компьютеры
- `POST /api/v1/bookings` - Создать бронирование
- `GET /api/v1/clients/{id}/bookings` - История бронирований
- `DELETE /api/v1/bookings/{id}` - Отменить бронирование

### Каталог и заказы
- `GET /api/v1/catalog` - Каталог товаров и услуг
- `POST /api/v1/orders` - Создать заказ
- `GET /api/v1/clients/{id}/orders` - История заказов

### Сессии и новости
- `GET /api/v1/clients/{id}/sessions` - История игровых сессий
- `GET /api/v1/news` - Новости и объявления клуба

## ⚙️ Конфигурация

### Переменные окружения

| Переменная | Описание | По умолчанию | Production |
|------------|----------|--------------|------------|
| `PORT` | Порт сервера | `8080` | `8080` |
| `DATABASE_PATH` | Путь к базе данных | `data/pcclub.db` | `/app/data/pcclub.db` |
| `JWT_SECRET` | Секрет для JWT токенов | `pcclub-local-dev-secret` | **Обязательно изменить!** |
| `ADMIN_API_TOKEN` | Токен администратора | - | **Обязательно задать!** |
| `READ_TIMEOUT` | Таймаут чтения запросов | `15s` | `30s` |
| `WRITE_TIMEOUT` | Таймаут записи ответов | `15s` | `30s` |
| `IDLE_TIMEOUT` | Таймаут простоя соединения | `60s` | `120s` |
| `MAX_HEADER_BYTES` | Макс. размер заголовков | `1MB` | `1MB` |
| `DB_MAX_OPEN_CONNS` | Макс. соединений с БД | `25` | `50` |
| `DB_MAX_IDLE_CONNS` | Макс. простаивающих соединений | `5` | `10` |
| `DB_CONN_MAX_LIFETIME` | Время жизни соединения | `5m` | `5m` |
| `DB_CONN_MAX_IDLE_TIME` | Время простоя соединения | `5m` | `5m` |
| `SEED_DEMO_DATA` | Загружать демо данные | `true` | `false` |

### Production конфигурация

Скопируйте `.env.example` в `.env` и настройте для production:

```bash
# Безопасность (ОБЯЗАТЕЛЬНО!)
JWT_SECRET=your-very-secure-jwt-secret-here-min-32-chars
ADMIN_API_TOKEN=your-secure-admin-token-here

# Производительность
READ_TIMEOUT=30s
WRITE_TIMEOUT=30s
IDLE_TIMEOUT=120s
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10

# Production режим
SEED_DEMO_DATA=false
```

## 🛡️ Безопасность и стабильность

### Middleware слои
- **Recovery** - Автоматическое восстановление от паник
- **Logging** - Детальное логирование всех запросов
- **Security Headers** - Заголовки безопасности (XSS, CSRF защита)
- **CORS** - Настроенная политика CORS для мобильных приложений
- **Rate Limiting** - Ограничение частоты запросов по IP
- **Timeout** - Таймауты для предотвращения зависших запросов

### Аутентификация
- JWT токены с автоматическим refresh механизмом
- Безопасное хранение паролей (bcrypt с солью)
- Rate limiting для auth эндпоинтов (защита от брутфорса)
- Валидация токенов на каждый запрос

### База данных
- Оптимизированный пул соединений SQLite
- Автоматическое управление жизненным циклом соединений
- Транзакционная безопасность для критических операций
- Graceful shutdown с корректным закрытием соединений

## 🚀 Развертывание

### Railway (рекомендуется)

1. Создайте GitHub репозиторий из этой папки
2. Подключите репозиторий в Railway
3. Настройте переменные окружения:

```bash
DATABASE_PATH=/app/data/pcclub.db
JWT_SECRET=your-secure-secret-here
ADMIN_API_TOKEN=your-admin-token-here
SEED_DEMO_DATA=false
READ_TIMEOUT=30s
WRITE_TIMEOUT=30s
DB_MAX_OPEN_CONNS=50
```

4. Railway автоматически настроит `PORT` и volume для данных

### Docker Production

```bash
# Только сервер
docker-compose up -d pcclub-server

# С Nginx reverse proxy (рекомендуется для production)
docker-compose --profile production up -d
```

### Обновление

```bash
git pull
docker-compose build
docker-compose up -d
```

## 🔧 Мониторинг

### Health Check

```bash
curl http://localhost:8080/health
# Ответ: {"status":"ok"}
```

### Логи производительности

Сервер автоматически логирует каждый запрос:
```
2026/05/21 15:02:25 GET /api/v1/pcs 200 15.2ms 127.0.0.1
2026/05/21 15:02:30 POST /api/v1/auth/login 200 45.1ms 192.168.1.100
```

### Просмотр логов

```bash
# Docker
docker-compose logs -f pcclub-server

# Локально - логи выводятся в stdout
```

## 📱 Совместимость с мобильным приложением

✅ **Полная совместимость с [PCCLUBMobile](https://github.com/Mtruyker/PCCLUBMobile):**

- API версионирование `/api/v1` 
- Стандартизированные ответы с обертками `data`/`result`
- JWT аутентификация с refresh токенами
- Все необходимые эндпоинты для мобильного приложения
- Правильные CORS настройки
- Консистентная обработка ошибок с кодами

### Формат ответов API v1

Успешные ответы:
```json
{
  "success": true,
  "data": { ... }
}
```

Ошибки:
```json
{
  "success": false,
  "message": "Описание ошибки",
  "code": "ERROR_CODE"
}
```

## 🔄 Legacy API

Сохранена полная обратная совместимость со старыми эндпоинтами:

### Авторизация
```http
POST /api/auth/register
POST /api/auth/login
```

### Административные методы
```http
GET /api/clients
POST /api/clients  
GET /api/computers
GET /api/sessions
GET /api/statistics
```

### Клиентские методы (требуют Authorization)
```http
GET /api/clients/{id}
PUT /api/clients/{id}
POST /api/clients/{id}/top-up
GET /api/clients/{id}/sessions
GET /api/clients/{id}/bookings
GET /api/clients/{id}/orders
```

## 🐛 Отладка

### Частые проблемы

1. **Сервер не запускается**
   ```bash
   # Проверьте логи
   docker-compose logs pcclub-server
   
   # Проверьте занятость порта
   netstat -tulpn | grep 8080
   ```

2. **Ошибки базы данных**
   ```bash
   # Проверьте права доступа к папке data/
   ls -la data/
   
   # Пересоздайте базу (ВНИМАНИЕ: потеря данных!)
   rm data/pcclub.db
   ```

3. **Проблемы с CORS**
   - Убедитесь что мобильное приложение использует правильный Origin
   - Проверьте настройки в `internal/app/http.go`

4. **JWT токены не работают**
   - Проверьте что `JWT_SECRET` одинаковый при перезапусках
   - Убедитесь что токен передается в заголовке `Authorization: Bearer <token>`

### Тестирование API

```bash
# Запуск тестов
go test ./...

# Сборка
go build ./cmd/server

# Проверка health check
curl http://localhost:8080/health

# Тест регистрации
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","phone":"1234567890","password":"test123"}'
```

## 📊 Схема базы данных

```sql
Clients(Id, Name, Phone, Email, PasswordHash, Balance)
Computers(Id, Name, Zone, X, Y, IsOccupied, HourPrice)  
Items(Id, Name, Category, Price, ImageUrl, Description)
News(Id, Title, Content, ImageUrl, PublishedAt)
Bookings(Id, ClientId, ClientName, PcId, ComputerName, StartTime, EndTime, TariffName, HourlyRate, DurationHours, TotalPrice, Status)
Orders(Id, ClientId, Date, TotalAmount, Status)
OrderItems(Id, OrderId, ProductId, ProductName, Quantity, Price)
OrderStatusHistory(Id, OrderId, Status, ChangedAt)
Sessions(Id, ClientId, ClientName, PcId, ComputerName, StartTime, EndTime, TotalCost, TariffName, HourlyRate)
Tariffs(Id, Name, CostPerHour, Description)
```

## 📞 Поддержка

- Создайте issue в GitHub репозитории для сообщения об ошибках
- Проверьте логи сервера для диагностики проблем
- Используйте health check эндпоинт для мониторинга

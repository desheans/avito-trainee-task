# Тестовое задание для стажёра Backend (осенняя волна 2025)

## Сервис назначения ревьюеров для Pull Request’ов

- Фреймворк: Echo v4 для HTTP API
- База данных: PostgreSQL с драйвером pgx/v5
- Контейнеризация: Docker с Makefile для автоматизации
- Тестирование: testcontainers-go для изолированных тестов БД + testify
- Документация: oapi-codegen для генерации кода из OpenAPI спецификации
- Миграции: golang-migrate для управления схемой БД
- Конфигурация: clearenv для загрузки переменных из .env файлов
- Логирование: slog для структурированного логирования
- Анализ кода: golangci-lint

### Инструкции к запуску

#### Клонирование репозитория
```bash
git clone https://github.com/desheans/avito-trainee-task.git
cd avito-trainee-task
```

#### Настройка проекта
```
make generate   # Генерация кода из OpenAPI
make lint       # Проверка кода линтером
```
Содержимое .env файла в корне проекта
```
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres_user
POSTGRES_PASSWORD=postgres_password
POSTGRES_DB=postgres_db

PG_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable

PORT=8080
ENV=dev
```

#### Запуск и остановка сервиса
```
make docker-up        # Запуск контейнеров
docker-compose up     # или
make docker-down      # Остановка контейнеров
```

#### Тестирование
```
make test-all           # Все тесты
make test-integration   # Интеграционные
make test-e2e           # E2E
```

### Настройка golangci-lint
#### Основные линтеры:
- bodyclose - проверяет закрытие HTTP response body
- gocyclo - анализирует цикломатическую сложность (мин. 15)
- misspell - ищет орфографические ошибки
- revive - современный линтер стиля кода
- sqlclosecheck - проверяет закрытие SQL соединений

### Примеры запросов
Добавление команды
```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true}
    ]
  }'
```
Создание Pull Request'а
```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search", 
    "author_id": "u1"
  }'
```

### Нюансы
- Для эндпоинта /users/getreview была добавлена обработка случая отсутствия пользователя - возвращается 404
```yaml
        "404":
          description: Пользователь не найден
          content:
            application/json:
              schema: { $ref: "#/components/schemas/ErrorResponse" }
```
- Для эндпоинта /users/stats была добавлена структура ответа, содержащая массив пользователей с количестом назначенных для них Pull Request'ов:
```yaml
    AssignmentCount:
      type: object
      required: [user_id, assignment_count]
      properties:
        user_id:
          type: string
        assignment_count:
          type: integer
    AssignmentCountStat:
      type: object
      required: [stats]
      properties:
        stats:
          type: array
          items:
            $ref: "#/components/schemas/AssignmentCount"
```
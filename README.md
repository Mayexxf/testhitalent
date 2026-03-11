# Organizational Structure API

API для управления организационной структурой компании с поддержкой иерархии подразделений и сотрудников.

## Описание

Приложение предоставляет REST API для:
- Управления иерархией подразделений (создание, обновление, удаление, получение)
- Управления сотрудниками в подразделениях
- Построения древовидной структуры подразделений с настраиваемой глубиной
- Каскадного удаления или переназначения сотрудников при удалении подразделений

## Технологический стек

- **Go 1.23** - язык программирования
- **net/http** - HTTP сервер
- **GORM** - ORM для работы с базой данных
- **PostgreSQL** - СУБД
- **goose** - миграции базы данных
- **Docker & docker-compose** - контейнеризация

## Быстрый старт

### Требования

- Docker
- docker-compose

### Запуск приложения

Запустите приложение с помощью docker-compose:
```bash
docker-compose up --build
```

Приложение будет доступно по адресу: `http://localhost:8080`

### Остановка приложения

```bash
docker-compose down
```

## API Endpoints

### 1. Создать подразделение
**POST** `/departments/`

Request:
```json
{
  "name": "Engineering",
  "parent_id": null
}
```

### 2. Получить подразделение
**GET** `/departments/{id}?depth=2&include_employees=true`

Query параметры:
- `depth` (int, default: 1, max: 5) - глубина вложенности
- `include_employees` (bool, default: true) - включать ли сотрудников

### 3. Обновить подразделение
**PATCH** `/departments/{id}`

Request:
```json
{
  "name": "New Name",
  "parent_id": 2
}
```

### 4. Удалить подразделение
**DELETE** `/departments/{id}?mode=cascade`

Query параметры:
- `mode` - `cascade` или `reassign`
- `reassign_to_department_id` - обязателен при mode=reassign

### 5. Создать сотрудника
**POST** `/departments/{id}/employees/`

Request:
```json
{
  "full_name": "Jane Smith",
  "position": "Product Manager",
  "hired_at": "2024-01-15"
}
```

## Тестирование

```bash
go test ./...
```

## Архитектура

Проект следует принципам чистой архитектуры с разделением на слои:
- **Handlers** - обработчики HTTP запросов
- **Service** - бизнес-логика
- **Repository** - работа с базой данных
- **Models** - структуры данных

## Структура проекта

```
.
├── cmd/api/              # Точка входа
├── internal/
│   ├── config/          # Конфигурация
│   ├── handlers/        # HTTP обработчики
│   ├── logger/          # Логирование
│   ├── models/          # Модели данных
│   ├── repository/      # Слой данных
│   └── service/         # Бизнес-логика
├── migrations/          # SQL миграции
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

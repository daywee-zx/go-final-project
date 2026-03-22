# Планировщик задач (Task Scheduler)

Веб-приложение для управления задачами и планирования с поддержкой повторяющихся событий. Финальный проект курса Go-developer-basic.

## Описание

Полнофункциональное веб-приложение для планирования задач с интеграцией REST API на Go. Приложение позволяет создавать, редактировать, удалять и отмечать выполненными задачи, а также устанавливать расписание повторения для автоматического переносения дат.

## Технологический стек

### Backend
- **Go 1.25.1** — язык программирования
- **SQLite** — база данных (modernc.org/sqlite)
- **sqlx** — работа с БД
- **http** — встроенный веб-сервер Go

### Frontend
- HTML5
- CSS3 с поддержкой тем
- JavaScript (Axios для API запросов)
- Отзывчивый дизайн

## Структура проекта

```
.
├── main.go                    # Точка входа приложения
├── go.mod                     # Зависимости проекта
├── Dockerfile                 # Docker конфигурация
├── README.md                  # Документация
├── logs.txt                   # Логи приложения
│
├── internal/                  # Внутренние пакеты
│   ├── server/               # HTTP сервер и обработчики
│   │   ├── server.go         # Инициализация сервера
│   │   ├── handlers.go       # HTTP обработчики API
│   │   ├── nextdate.go       # Расчет следующей даты повтора
│   │   └── utils.go          # Утилиты
│   │
│   └── scheduler_db/         # Работа с БД
│       └── scheduler_db.go   # CRUD операции с задачами
│
├── web/                       # Статические файлы и интерфейс
│   ├── index.html            # Главная страница
│   ├── login.html            # Страница входа
│   ├── css/
│   │   ├── style.css         # Основные стили
│   │   └── theme.css         # Темы оформления
│   └── js/
│       ├── axios.min.js      # HTTP клиент
│       └── scripts.min.js    # Логика приложения
│
└── tests/                     # Тесты
    ├── app_1_test.go
    ├── nextdate_3_test.go
    ├── db_2_test.go
    ├── addtask_4_test.go
    ├── tasks_5_test.go
    ├── task_6_test.go
    ├── task_7_test.go
    └── settings.go
```

## Быстрый старт

### Требования

- Go 1.25 или выше
- Git

### Установка и запуск

1. **Клонирование репозитория**
   ```bash
   git clone <repository-url>
   cd go-final-project
   ```

2. **Скачивание зависимостей**
   ```bash
   go mod download
   ```

3. **Запуск приложения**
   ```bash
   go run main.go
   ```

   Приложение запустится на `http://localhost:7540`

### Переменные окружения

Приложение поддерживает следующие переменные окружения:

| Переменная | Описание | Значение по умолчанию |
|-----------|---------|----------------------|
| `TODO_PORT` | Порт сервера | `7540` |
| `TODO_DBFILE` | Путь к файлу БД | `scheduler.db` |
| `TODO_PASSWORD` | Пароль для доступа | `123456` |

**Пример запуска с переменными окружения:**
```bash
TODO_PORT=8080 TODO_DBFILE=/data/tasks.db TODO_PASSWORD=mypassword go run main.go
```

## Docker

### Сборка Docker образа

```bash
docker build -t task-scheduler .
```

### Запуск контейнера

```bash
docker run -p 7540:7540 \
  -e TODO_PORT=7540 \
  -e TODO_PASSWORD=mypassword \
  -v /data:/app/data \
  task-scheduler
```

## Схема базы данных

```sql
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title TEXT NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat CHAR(128) NOT NULL DEFAULT ""
);
```

| Поле | Тип | Описание |
|------|-----|---------|
| `id` | INTEGER | Уникальный идентификатор (автоинкремент) |
| `date` | CHAR(8) | Дата в формате YYYYMMDD |
| `title` | TEXT | Название задачи |
| `comment` | TEXT | Описание задачи |
| `repeat` | CHAR(128) | Правило повтора задачи |

## Логирование

Логи приложения сохраняются в файл `logs.txt`. Если файл невозможно создать, логи выводятся в `stdout`.

## Безопасность

- Пароль хранится в виде SHA256 хеша
- Аутентификация требуется для большинства операций с задачами

## Лицензия

Этот проект является учебным проектом курса Go-developer-basic.

## Автор

Разработано как финальный проект студента курса Go-developer-basic.

## Использование ИИ

В процессе разработки ИИ был использован исключительно для тестирования и написания этого README
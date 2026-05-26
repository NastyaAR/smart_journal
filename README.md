# Smart Journal - Учебный журнал с AI-рекомендациями

## Быстрый старт

### 1. Настройка переменных окружения

```bash
cp .env.example .env
```

Откройте `.env` и укажите:
- `GROQ_API_KEY` - ваш API ключ от Groq (получить на https://console.groq.com/keys)

### 2. Запуск Docker

```bash
docker-compose up --build
```

Сервисы будут доступны по адресам:
- **Frontend**: http://localhost:5173
- **API**: http://localhost:3000
- **Swagger UI**: http://localhost:8081
- **PostgreSQL**: localhost:5432

### 3. Вход в систему

**Учитель** (демо-аккаунты из миграции):
- Email: `anna.ivanova@school.edu`
- Пароль: `password`

**Студент** (демо-аккаунты из миграции):
- Email: `ivan.petrov@student.edu`
- Пароль: `password`

Или зарегистрируйте нового студента через форму регистрации.

---

## Архитектура проекта

```
smart_journal-main/
├── ai/                      # AI сервис (Python/FastAPI)
│   ├── main.py             # Генерация рекомендаций через Groq
│   ├── requirements.txt
│   └── Dockerfile
├── cmd/api/                 # Backend API (Go/Fiber)
│   └── main.go
├── frontend/                # Frontend (React/Vite/TypeScript)
│   └── src/
├── internal/                # Go бизнес-логика
│   ├── handlers/           # HTTP обработчики
│   ├── services/           # Бизнес-сервисы
│   ├── repositories/       # Работа с БД
│   └── models/             # Модели данных
├── migrations/              # Миграции БД (goose)
├── docker-compose.yml       # Docker конфигурация
└── .env.example            # Шаблон переменных окружения
```

---

## Компоненты

### AI Сервис (Python)
- **Порт**: 8000
- **Фреймворк**: FastAPI
- **Модель**: Llama 3.1 8B через Groq API
- **Эндпоинты**:
  - `GET /` - проверка статуса
  - `POST /get_recommendations` - генерация рекомендаций

### Backend API (Go)
- **Порт**: 3000
- **Фреймворк**: Fiber
- **База данных**: PostgreSQL
- **Основные функции**:
  - Аутентификация (сессии)
  - Управление студентами и группами
  - Оценки и предметы
  - Активности и достижения
  - Токены AMT (блокчейн)
  - AI-рекомендации

### Frontend (React)
- **Порт**: 5173
- **Стек**: React 18, TypeScript, Vite
- **Роли**: Учитель, Студент

---

## Решение проблем

### AI рекомендации не работают
1. Проверьте что `GROQ_API_KEY` указан в `.env`
2. Убедитесь что AI сервис запущен: `docker-compose ps llm`
3. Проверьте логи: `docker-compose logs llm`

### Ошибки базы данных
1. Перезапустите миграции: `docker-compose up migrations`
2. Проверьте статус: `docker-compose logs migrations`

### Студенты не загружаются
Новая миграция `006_seed_students.sql` автоматически загружает:
- 10 студентов в 3 группах (БПИ-231, БПИ-232, БПИ-233)
- Тестовые оценки по 3 предметам
- Стартовые токены (100 AMT)

---

## Демо-данные

### Группы
- БПИ-231, БПИ-232, БПИ-233

### Предметы
- Математика, Программирование, Базы данных, Алгоритмы, Веб-разработка

### Товары мерча
- Доступ к курсу по Python (50 AMT)
- Футболка с логотипом (30 AMT)
- Онлайн-встреча с ментором (40 AMT)
- И другие

---

## Разработка

### Локальный запуск API (Go)
```bash
cd cmd/api
go run main.go
```

### Локальный запуск AI (Python)
```bash
cd ai
pip install -r requirements.txt
uvicorn main:app --reload
```

### Локальный запуск Frontend
```bash
cd frontend
npm install
npm run dev
```

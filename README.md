# EduChain AI - Учебный журнал с AI-рекомендациями

**EduChain AI** — это гибридная платформа для образовательных учреждений, которая объединяет:
- централизованный бэкенд (Go + PostgreSQL) для ведения журнала оценок и управления пользователями;
- смарт-контракт в сети Ethereum (локально Ganache) для токенизации учебных достижений (токен AMT, ERC‑20);
- ИИ‑сервис на Python (FastAPI) для генерации персонализированных рекомендаций курсов на основе успеваемости студента.

Система автоматически начисляет токены за подтверждённые достижения, а студенты могут тратить токены на реальные привилегии. Блокчейн обеспечивает неизменяемый аудит всех начислений, исключая манипуляции.

## Основные возможности

### Для студентов
- Просмотр своих оценок и баланса токенов AMT
- Автоматическое получение токенов за достижения
- Подача заявок на подтверждение внеучебных достижений (конференции, олимпиады, публикации)
- Получение персонализированных ИИ‑рекомендаций дополнительных курсов 
- Покупка мерча за токены

### Для преподавателей
- Выставление оценок
- Подтверждение или отклонение заявок на достижения студентов
- Просмотр успеваемости группы и истории начислений токенов
- Ручное начисление токенов за особые заслуги
- Прозрачный аудит всех наград через смарт‑контракт

## Токен AMT (ERC‑20)

### Функции смарт‑контракта
- `awardStudent(address student, uint256 amount, string reason)` — начисление токенов студенту (только преподаватель)
- `redeem(uint256 amount)` — сжигание токенов для получения привилегий
- `getBalance(address student)` — проверка баланса токенов

### Токеномика
- Автоматическое начисление токенов за подтверждённые достижения
- Обмен токенов на мерч
- Дефляционная механика: токены сжигаются при вызове `redeem()`, общее предложение уменьшается

## Быстрый старт

### 1. Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```bash
cp .env.example .env
```

Откройте `.env` и укажите:
- `OPENROUTER_API_KEY` - ваш API ключ от OpenRouter (получить на https://openrouter.ai/keys)

**Пример .env:**
```env
# Database
PSQL_DB=smartjournal
PSQL_USER=admin
PSQL_PASSWORD=12345
DATABASE_URL=postgres://admin:12345@localhost:5432/smartjournal?sslmode=disable
DATABASE_URL_API=postgres://admin:12345@postgres_container:5432/smartjournal?sslmode=disable
PSQL_PORT=5432

# Blockchain
CONTRACT_ADDRESS=0xYourContractAddress
RPC_URL=https://sepolia.infura.io/v3/YOUR_PROJECT_ID
CONTRACT_ADMIN_PRIVATE_KEY=your_contract_admin_private_key

# LLM Service - OpenRouter (бесплатные модели)
OPENROUTER_API_KEY=sk-or-ваш_ключ
AI_SERVICE_URL=http://llm:8000

# App
PORT=3000
VITE_BACKEND_URL=http://host.docker.internal:3000
```

### 2. Запуск Docker

```powershell
docker compose up --build
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
│   ├── main.py             # Генерация рекомендаций через OpenRouter
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
- **Модель**: Mistral 7B Instruct через OpenRouter
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

1. **Проверьте что OPENROUTER_API_KEY указан в .env**
   ```powershell
   Get-Content .env | Select-String OPENROUTER
   ```

2. **Получите новый ключ на https://openrouter.ai/keys**
   - Зарегистрируйтесь через Google/GitHub
   - Создайте API ключ
   - Скопируйте и вставьте в .env

3. **Перезапустите сервисы**
   ```powershell
   docker compose down
   docker compose up --build -d
   ```

4. **Проверьте логи AI сервиса**
   ```powershell
   docker compose logs llm
   ```

5. **Протестируйте AI сервис напрямую**
   ```powershell
   # Проверка статуса
   Invoke-RestMethod http://localhost:8000/
   
   # Тест генерации
   $body = @{
       student_id = "2"
       student_name = "Иван"
       student_surname = "Иванов"
       grades = @(@{subject = "Математика"; score = 85})
   } | ConvertTo-Json -Depth 5
   
   $result = Invoke-RestMethod http://localhost:8000/get_recommendations `
       -Method Post `
       -ContentType "application/json" `
       -Body $body
   
   # Сохранить результат в файл (для просмотра русского текста)
   $result | ConvertTo-Json -Depth 10 | Out-File result.json -Encoding UTF8
   notepad result.json
   ```

### Ошибки базы данных

1. Перезапустите миграции:
   ```powershell
   docker compose up migrations
   ```

2. Проверьте статус:
   ```powershell
   docker compose logs migrations
   ```

### Студенты не загружаются

Миграция `006_seed_students.sql` автоматически загружает:
- 10 студентов в 3 группах (БПИ-231, БПИ-232, БПИ-233)
- Тестовые оценки по 3 предметам
- Стартовые токены (100 AMT)

Проверьте применение миграций:
```powershell
docker compose logs migrations
```

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
python main.py
```

### Локальный запуск Frontend
```bash
cd frontend
npm install
npm run dev
```

---

## API Ключи

### OpenRouter (бесплатно)
1. Зайдите на https://openrouter.ai/keys
2. Войдите через Google/GitHub
3. Создайте API ключ
4. Используйте бесплатные модели:
   - `mistralai/mistral-7b-instruct-v0.1`
   - `google/gemma-7b-it:free`
   - `meta-llama/llama-3-8b-instruct:free`

---

## Структура миграций

```
migrations/
├── 001_init.sql                    # Базовая схема
├── 002_seed_teachers_and_merch.sql # Учителя и мерч
├── 003_activity_status_...         # Статусы активностей
├── 004_student_recommendations.sql # Таблица рекомендаций
├── 005_grade_dates_and_...         # Даты оценок
└── 006_seed_students.sql           # Тестовые студенты
```

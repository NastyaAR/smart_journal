# Curl-запросы по юзкейсам

Ниже примеры для локального сервера `http://localhost:3000`.

```bash
BASE_URL=http://localhost:3000
TEACHER_COOKIES=./cookies_teacher.txt
STUDENT_COOKIES=./cookies_student.txt
```

## Health check

```bash
curl "$BASE_URL/health"
```

## Учитель

### 1. Авторизоваться

Сидовые учителя создаются миграцией `002_seed_teachers_and_merch.sql`. Пароль: `password`.

```bash
curl -i -c "$TEACHER_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"login":"anna.ivanova@school.edu","password":"password"}' \
  "$BASE_URL/auth/login"
```

Проверить сессию:

```bash
curl -b "$TEACHER_COOKIES" "$BASE_URL/auth/session"
```

### 2. Создать группу

Созданная группа автоматически привязывается к текущему учителю.

```bash
curl -i -b "$TEACHER_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"name":"БПИ-231"}' \
  "$BASE_URL/teachers/groups"
```

### 3. Посмотреть свои группы

```bash
curl -b "$TEACHER_COOKIES" "$BASE_URL/teachers/groups"
```

### 4. Привязать существующую группу к учителю

```bash
curl -i -b "$TEACHER_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"group_id":1}' \
  "$BASE_URL/teachers/groups/attach"
```

### 5. Создать предмет

Созданный предмет автоматически привязывается к текущему учителю.

```bash
curl -i -b "$TEACHER_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"name":"Математика"}' \
  "$BASE_URL/teachers/subjects"
```

### 6. Привязать существующий предмет к учителю

```bash
curl -i -b "$TEACHER_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"subject_id":1}' \
  "$BASE_URL/teachers/subjects/attach"
```

### 7. Добавить зарегистрированного ученика в группу

```bash
curl -i -b "$TEACHER_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"student_id":1,"group_id":1}' \
  "$BASE_URL/teachers/groups/add-student"
```

### 8. Поставить оценку

```bash
curl -i -b "$TEACHER_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"student_id":1,"subject_id":1,"value":5}' \
  "$BASE_URL/teachers/grades"
```

### 9. Посмотреть оценки группы

```bash
curl -b "$TEACHER_COOKIES" "$BASE_URL/teachers/groups/1/grades"
```

### 10. Посмотреть pending-заявки на активности

Учитель увидит только заявки учеников из привязанных к нему групп.

```bash
curl -b "$TEACHER_COOKIES" "$BASE_URL/teachers/achievements/pending"
```

### 11. Подтвердить активность и начислить токены

```bash
curl -i -b "$TEACHER_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"achievement_id":1}' \
  "$BASE_URL/teachers/achievements/confirm"
```

### 12. Отклонить активность

```bash
curl -i -b "$TEACHER_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"achievement_id":1}' \
  "$BASE_URL/teachers/achievements/deny"
```

### 13. Начислить токены вручную

```bash
curl -i -b "$TEACHER_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"student_id":1,"amount":10}' \
  "$BASE_URL/teachers/tokens/award"
```

### 14. Выйти

```bash
curl -i -b "$TEACHER_COOKIES" -c "$TEACHER_COOKIES" \
  -X POST "$BASE_URL/auth/logout"
```

## Ученик

### 1. Зарегистрироваться

`group_id` можно передать `0`, если учитель добавит ученика в группу позже.

```bash
curl -i \
  -H "Content-Type: application/json" \
  -d '{"name":"Иван Петров","email":"ivan.petrov@example.com","group_id":0,"password":"secret"}' \
  "$BASE_URL/register"
```

### 2. Авторизоваться

```bash
curl -i -c "$STUDENT_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"login":"ivan.petrov@example.com","password":"secret"}' \
  "$BASE_URL/auth/login"
```

Проверить сессию:

```bash
curl -b "$STUDENT_COOKIES" "$BASE_URL/auth/session"
```

### 3. Посмотреть свою группу

```bash
curl -b "$STUDENT_COOKIES" "$BASE_URL/students/group"
```

### 4. Посмотреть оценки своей группы

```bash
curl -b "$STUDENT_COOKIES" "$BASE_URL/students/grades"
```

### 5. Посмотреть баланс токенов

```bash
curl -b "$STUDENT_COOKIES" "$BASE_URL/students/balance"
```

### 6. Посмотреть доступный мерч

```bash
curl -b "$STUDENT_COOKIES" "$BASE_URL/students/merch"
```

### 7. Купить мерч за токены

```bash
curl -i -b "$STUDENT_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"merch_id":1}' \
  "$BASE_URL/students/merch/buy"
```

### 8. Посмотреть историю покупок

```bash
curl -b "$STUDENT_COOKIES" "$BASE_URL/students/purchases"
```

### 9. Отправить дополнительную активность на подтверждение

```bash
curl -i -b "$STUDENT_COOKIES" \
  -H "Content-Type: application/json" \
  -d '{"title":"Участие в олимпиаде","description":"Решил региональный этап"}' \
  "$BASE_URL/students/achievements"
```

### 10. Посмотреть свои активности

```bash
curl -b "$STUDENT_COOKIES" "$BASE_URL/students/achievements"
```

### 11. Выйти

```bash
curl -i -b "$STUDENT_COOKIES" -c "$STUDENT_COOKIES" \
  -X POST "$BASE_URL/auth/logout"
```

## Мини-сценарий end-to-end

```bash
BASE_URL=http://localhost:3000
TEACHER_COOKIES=./cookies_teacher.txt
STUDENT_COOKIES=./cookies_student.txt

curl -i -c "$TEACHER_COOKIES" -H "Content-Type: application/json" \
  -d '{"login":"anna.ivanova@school.edu","password":"password"}' \
  "$BASE_URL/auth/login"

curl -i -b "$TEACHER_COOKIES" -H "Content-Type: application/json" \
  -d '{"name":"БПИ-231"}' \
  "$BASE_URL/teachers/groups"

curl -i -b "$TEACHER_COOKIES" -H "Content-Type: application/json" \
  -d '{"name":"Математика"}' \
  "$BASE_URL/teachers/subjects"

curl -i -H "Content-Type: application/json" \
  -d '{"name":"Иван Петров","email":"ivan.petrov@example.com","group_id":0,"password":"secret"}' \
  "$BASE_URL/register"

curl -i -c "$STUDENT_COOKIES" -H "Content-Type: application/json" \
  -d '{"login":"ivan.petrov@example.com","password":"secret"}' \
  "$BASE_URL/auth/login"

curl -i -b "$TEACHER_COOKIES" -H "Content-Type: application/json" \
  -d '{"student_id":1,"group_id":1}' \
  "$BASE_URL/teachers/groups/add-student"

curl -i -b "$TEACHER_COOKIES" -H "Content-Type: application/json" \
  -d '{"student_id":1,"subject_id":1,"value":5}' \
  "$BASE_URL/teachers/grades"

curl -i -b "$STUDENT_COOKIES" -H "Content-Type: application/json" \
  -d '{"title":"Участие в олимпиаде","description":"Решил региональный этап"}' \
  "$BASE_URL/students/achievements"

curl -b "$TEACHER_COOKIES" "$BASE_URL/teachers/achievements/pending"

curl -i -b "$TEACHER_COOKIES" -H "Content-Type: application/json" \
  -d '{"achievement_id":1}' \
  "$BASE_URL/teachers/achievements/confirm"

curl -b "$STUDENT_COOKIES" "$BASE_URL/students/balance"
curl -b "$STUDENT_COOKIES" "$BASE_URL/students/grades"
```

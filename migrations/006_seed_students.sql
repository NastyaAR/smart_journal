-- +goose Up
-- Создаем группы для студентов
INSERT INTO groups (name) VALUES
('БПИ-231'),
('БПИ-232'),
('БПИ-233')
ON CONFLICT DO NOTHING;

-- Создаем предметы
INSERT INTO subjects (name) VALUES
('Математика'),
('Программирование'),
('Базы данных'),
('Алгоритмы'),
('Веб-разработка')
ON CONFLICT DO NOTHING;

-- Пароль для всех студентов: password
WITH student_seed(name, email, password, group_name) AS (
    VALUES
    ('Иван Петров', 'ivan.petrov@student.edu', 'password', 'БПИ-231'),
    ('Мария Сидорова', 'maria.sidorova@student.edu', 'password', 'БПИ-231'),
    ('Алексей Иванов', 'alexey.ivanov@student.edu', 'password', 'БПИ-231'),
    ('Елена Козлова', 'elena.kozlova@student.edu', 'password', 'БПИ-232'),
    ('Дмитрий Морозов', 'dmitry.morozov@student.edu', 'password', 'БПИ-232'),
    ('Ольга Новикова', 'olga.novikova@student.edu', 'password', 'БПИ-232'),
    ('Сергей Лебедев', 'sergey.lebedev@student.edu', 'password', 'БПИ-233'),
    ('Татьяна Григорьева', 'tatyana.grigoryeva@student.edu', 'password', 'БПИ-233'),
    ('Николай Попов', 'nikolay.popov@student.edu', 'password', 'БПИ-233'),
    ('Анна Васильева', 'anna.vasilyeva@student.edu', 'password', 'БПИ-233')
),
inserted_users AS (
    INSERT INTO users (login, password, role)
    SELECT email, password, 'student'
    FROM student_seed
    ON CONFLICT (login) DO NOTHING
    RETURNING id, login
),
student_users AS (
    SELECT id, login FROM inserted_users
    UNION
    SELECT u.id, u.login
    FROM users u
    JOIN student_seed ts ON ts.email = u.login
    WHERE u.role = 'student'
),
group_ids AS (
    SELECT name, id FROM groups
),
inserted_students AS (
    INSERT INTO students (name, email, group_id, user_id, tokens)
    SELECT ss.name, ss.email, gi.id, su.id, 100
    FROM student_seed ss
    JOIN student_users su ON su.login = ss.email
    JOIN group_ids gi ON gi.name = ss.group_name
    ON CONFLICT (email) DO UPDATE SET
        name = EXCLUDED.name,
        group_id = EXCLUDED.group_id,
        tokens = 100
    RETURNING id, email
)
SELECT * FROM inserted_students;

-- Добавляем тестовые оценки для студентов
WITH student_ids AS (
    SELECT id, email FROM students
),
subject_ids AS (
    SELECT id, name FROM subjects
),
grade_data(email, subject_name, value) AS (
    VALUES
    ('ivan.petrov@student.edu', 'Математика', 85),
    ('ivan.petrov@student.edu', 'Программирование', 92),
    ('ivan.petrov@student.edu', 'Базы данных', 78),
    ('maria.sidorova@student.edu', 'Математика', 95),
    ('maria.sidorova@student.edu', 'Программирование', 88),
    ('maria.sidorova@student.edu', 'Базы данных', 91),
    ('alexey.ivanov@student.edu', 'Математика', 72),
    ('alexey.ivanov@student.edu', 'Программирование', 65),
    ('alexey.ivanov@student.edu', 'Базы данных', 70),
    ('elena.kozlova@student.edu', 'Математика', 88),
    ('elena.kozlova@student.edu', 'Программирование', 95),
    ('elena.kozlova@student.edu', 'Базы данных', 82),
    ('dmitry.morozov@student.edu', 'Математика', 60),
    ('dmitry.morozov@student.edu', 'Программирование', 55),
    ('dmitry.morozov@student.edu', 'Базы данных', 65)
)
INSERT INTO grades (student_id, subject_id, value, lesson_date)
SELECT sid.id, subj.id, gd.value, CURRENT_DATE
FROM grade_data gd
JOIN student_ids sid ON sid.email = gd.email
JOIN subject_ids subj ON subj.name = gd.subject_name
ON CONFLICT DO NOTHING;

-- +goose Down
-- Удаляем оценки студентов
DELETE FROM grades WHERE student_id IN (
    SELECT s.id FROM students s
    JOIN users u ON s.user_id = u.id
    WHERE u.role = 'student' AND u.login LIKE '%@student.edu'
);

-- Удаляем студентов
DELETE FROM students WHERE email LIKE '%@student.edu';

-- Удаляем пользователей-студентов
DELETE FROM users WHERE login LIKE '%@student.edu' AND role = 'student';

-- Удаляем группы (если они были созданы только для этого сида)
DELETE FROM groups WHERE name IN ('БПИ-231', 'БПИ-232', 'БПИ-233');

-- Удаляем предметы (если они были созданы только для этого сида)
DELETE FROM subjects WHERE name IN ('Математика', 'Программирование', 'Базы данных', 'Алгоритмы', 'Веб-разработка');

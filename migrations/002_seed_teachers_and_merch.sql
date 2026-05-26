-- +goose Up
-- Пароль для всех учителей: password
WITH teacher_seed(name, email, password) AS (
    VALUES
    ('Анна Иванова', 'anna.ivanova@school.edu', 'password'),
    ('Сергей Петров', 'sergey.petrov@school.edu', 'password'),
    ('Мария Сидорова', 'maria.sidorova@school.edu', 'password'),
    ('Дмитрий Козлов', 'dmitry.kozlov@school.edu', 'password'),
    ('Елена Морозова', 'elena.morozova@school.edu', 'password'),
    ('Алексей Новиков', 'alexey.novikov@school.edu', 'password'),
    ('Ольга Лебедева', 'olga.lebedeva@school.edu', 'password'),
    ('Игорь Васильев', 'igor.vasiliev@school.edu', 'password'),
    ('Татьяна Григорьева', 'tatyana.grigoryeva@school.edu', 'password'),
    ('Николай Попов', 'nikolay.popov@school.edu', 'password')
),
inserted_users AS (
    INSERT INTO users (login, password, role)
    SELECT email, password, 'teacher'
    FROM teacher_seed
    ON CONFLICT (login) DO NOTHING
    RETURNING id, login
),
teacher_users AS (
    SELECT id, login FROM inserted_users
    UNION
    SELECT u.id, u.login
    FROM users u
    JOIN teacher_seed ts ON ts.email = u.login
    WHERE u.role = 'teacher'
)
INSERT INTO teachers (name, email, user_id)
SELECT ts.name, ts.email, u.id
FROM teacher_seed ts
JOIN teacher_users u ON u.login = ts.email
WHERE NOT EXISTS (
    SELECT 1 FROM teachers t WHERE t.user_id = u.id
);

-- Добавляем товары мерча
INSERT INTO merch (title, description, price) VALUES
('Доступ к курсу по Python', '6-месячный доступ к продвинутому курсу по Python', 50),
('Футболка с логотипом школы', 'Хлопковая футболка, размеры S-XXL', 30),
('Онлайн-встреча с ментором', 'Персональная консультация на 30 минут', 40),
('Набор наклеек и ручек', 'Сувенирный набор от школы', 15),
('Бесплатное участие в хакатоне', 'Пропуск на внутренний хакатон', 25),
('VIP-доступ к вебинарам', 'Просмотр всех записей вебинаров без ограничений', 60),
('Электронный сертификат', 'Оформленный PDF-сертификат об успехах', 20);

-- +goose Down
DELETE FROM teachers WHERE email LIKE '%@school.edu';
DELETE FROM users WHERE login LIKE '%@school.edu' AND role = 'teacher';
DELETE FROM merch WHERE price > 0;

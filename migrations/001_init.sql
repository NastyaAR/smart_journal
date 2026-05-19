-- +goose Up
-- Создаем таблицу пользователей для аутентификации
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('student', 'teacher'))
);

-- Создаем таблицу учителей
CREATE TABLE teachers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE
);

-- Создаем таблицу групп
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- Создаем таблицу студентов
CREATE TABLE students (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    group_id INTEGER REFERENCES groups(id),
    tokens INTEGER DEFAULT 0,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE
);

-- Создаем таблицу предметов
CREATE TABLE subjects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- Создаем таблицу оценок
CREATE TABLE grades (
    id SERIAL PRIMARY KEY,
    student_id INTEGER REFERENCES students(id),
    subject_id INTEGER REFERENCES subjects(id),
    value INTEGER NOT NULL
);

-- Создаем таблицу достижений
CREATE TABLE achievements (
    id SERIAL PRIMARY KEY,
    student_id INTEGER REFERENCES students(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    confirmed BOOLEAN DEFAULT false,
    confirmed_by_teacher_id INTEGER REFERENCES teachers(id)
);

-- Создаем таблицу мерча
CREATE TABLE merch (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    price INTEGER NOT NULL
);

-- +goose Down
-- Удаляем все таблицы
DROP TABLE IF EXISTS users, teachers, groups, students, subjects, grades, achievements, merch CASCADE;
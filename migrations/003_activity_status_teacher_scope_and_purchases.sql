-- +goose Up
ALTER TABLE achievements
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'confirmed', 'denied'));

UPDATE achievements
SET status = CASE
    WHEN confirmed = true THEN 'confirmed'
    WHEN confirmed = false AND confirmed_by_teacher_id IS NOT NULL THEN 'denied'
    ELSE 'pending'
END;

CREATE TABLE teacher_groups (
    teacher_id INTEGER NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    PRIMARY KEY (teacher_id, group_id)
);

CREATE TABLE teacher_subjects (
    teacher_id INTEGER NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    subject_id INTEGER NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    PRIMARY KEY (teacher_id, subject_id)
);

CREATE TABLE purchases (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    merch_id INTEGER NOT NULL REFERENCES merch(id),
    price INTEGER NOT NULL CHECK (price > 0),
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS purchases;
DROP TABLE IF EXISTS teacher_subjects;
DROP TABLE IF EXISTS teacher_groups;
ALTER TABLE achievements DROP COLUMN IF EXISTS status;

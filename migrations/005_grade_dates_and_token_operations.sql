-- +goose Up
ALTER TABLE grades
    ADD COLUMN lesson_date DATE NOT NULL DEFAULT CURRENT_DATE,
    ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT now();

CREATE TABLE token_operations (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    teacher_id INTEGER REFERENCES teachers(id) ON DELETE SET NULL,
    amount INTEGER NOT NULL,
    operation_type VARCHAR(32) NOT NULL CHECK (operation_type IN ('achievement_reward', 'manual_award', 'purchase')),
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_token_operations_student_created
    ON token_operations (student_id, created_at DESC, id DESC);

CREATE INDEX idx_token_operations_teacher_created
    ON token_operations (teacher_id, created_at DESC, id DESC);

INSERT INTO token_operations (student_id, teacher_id, amount, operation_type, reason, created_at)
SELECT
    a.student_id,
    a.confirmed_by_teacher_id,
    10,
    'achievement_reward',
    a.title,
    now()
FROM achievements a
WHERE a.status = 'confirmed' OR a.confirmed = true;

INSERT INTO token_operations (student_id, amount, operation_type, reason, created_at)
SELECT
    p.student_id,
    -p.price,
    'purchase',
    COALESCE(m.title, 'Purchase #' || p.id::text),
    p.created_at
FROM purchases p
LEFT JOIN merch m ON m.id = p.merch_id;

-- +goose Down
DROP TABLE IF EXISTS token_operations;

ALTER TABLE grades
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS lesson_date;

-- +goose Up
CREATE TABLE student_recommendations (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_student_recommendations_student_created
    ON student_recommendations (student_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS student_recommendations;

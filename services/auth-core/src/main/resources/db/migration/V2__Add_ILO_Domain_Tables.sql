
-- Create domains table
CREATE TABLE IF NOT EXISTS ilo_domains (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT
);

-- Create evaluation levels table
CREATE TABLE IF NOT EXISTS ilo_levels (
    id BIGSERIAL PRIMARY KEY,
    min_percent INTEGER NOT NULL,
    max_percent INTEGER NOT NULL,
    level_name VARCHAR(255) NOT NULL,
    suggestion TEXT,
    CONSTRAINT ilo_level_range_check CHECK (min_percent <= max_percent)
);

-- Create mapping table for questions to domains
CREATE TABLE IF NOT EXISTS ilo_question_domains (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES ilo_questions(id),
    domain_id BIGINT NOT NULL REFERENCES ilo_domains(id),
    weight NUMERIC(5, 2) NOT NULL DEFAULT 1.0,
    CONSTRAINT ilo_question_domain_unique UNIQUE (question_id, domain_id)
);

-- Create domain scores table for test results
CREATE TABLE IF NOT EXISTS ilo_domain_scores (
    id BIGSERIAL PRIMARY KEY,
    test_result_id BIGINT NOT NULL REFERENCES ilo_test_results(id),
    domain_id BIGINT NOT NULL REFERENCES ilo_domains(id),
    raw_score INTEGER NOT NULL,
    percent_score REAL NOT NULL,
    level_id BIGINT REFERENCES ilo_levels(id),
    rank INTEGER,
    CONSTRAINT ilo_domain_score_unique UNIQUE (test_result_id, domain_id)
);

-- Create individual answers table
CREATE TABLE IF NOT EXISTS ilo_answers (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    question_id BIGINT NOT NULL REFERENCES ilo_questions(id),
    selected_option INTEGER NOT NULL,
    test_result_id BIGINT NOT NULL REFERENCES ilo_test_results(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create career mapping table
CREATE TABLE IF NOT EXISTS ilo_career_map (
    id BIGSERIAL PRIMARY KEY,
    domain_id BIGINT NOT NULL REFERENCES ilo_domains(id),
    career_field VARCHAR(255) NOT NULL,
    description TEXT,
    priority INTEGER DEFAULT 0
);

-- Add suggested_careers column to test results
ALTER TABLE ilo_test_results
ADD COLUMN IF NOT EXISTS suggested_careers TEXT;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_ilo_questions_domain ON ilo_question_domains (domain_id);
CREATE INDEX IF NOT EXISTS idx_ilo_domain_scores_test_result ON ilo_domain_scores (test_result_id);
CREATE INDEX IF NOT EXISTS idx_ilo_answers_test_result ON ilo_answers (test_result_id);
CREATE INDEX IF NOT EXISTS idx_ilo_answers_user ON ilo_answers (user_id);
CREATE INDEX IF NOT EXISTS idx_ilo_career_domain ON ilo_career_map (domain_id);

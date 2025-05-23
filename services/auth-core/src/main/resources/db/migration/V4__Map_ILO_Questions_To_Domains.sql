-- Helper function to map question numbers to domains
CREATE OR REPLACE FUNCTION map_question_to_domain(question_number INTEGER)
RETURNS TEXT AS $$
BEGIN
    -- Map questions to domains based on ILO test structure (12 questions per domain)
    -- Questions 1-12: Language skills
    IF question_number BETWEEN 1 AND 12 THEN
        RETURN 'LANG';
    -- Questions 13-24: Logic and analytical skills  
    ELSIF question_number BETWEEN 13 AND 24 THEN
        RETURN 'LOGIC';
    -- Questions 25-36: Design and creative skills
    ELSIF question_number BETWEEN 25 AND 36 THEN
        RETURN 'DESIGN';
    -- Questions 37-48: People and interpersonal skills
    ELSIF question_number BETWEEN 37 AND 48 THEN
        RETURN 'PEOPLE';
    -- Questions 49-60: Mechanical and physical skills
    ELSIF question_number BETWEEN 49 AND 60 THEN
        RETURN 'MECH';
    ELSE
        -- Default to LANG for any unexpected question numbers
        RETURN 'LANG';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Ensure unique mapping between question and domain
ALTER TABLE ilo_question_domains
    ADD CONSTRAINT uq_question_domain UNIQUE (question_id, domain_id);

-- Map existing questions to domains based on question numbers
INSERT INTO ilo_question_domains (question_id, domain_id, weight)
SELECT 
    q.id AS question_id,
    d.id AS domain_id,
    1.0 AS weight -- Default weight
FROM 
    ilo_questions q
JOIN 
    ilo_domains d ON d.code = map_question_to_domain(q.question_number)
ON CONFLICT (question_id, domain_id) DO NOTHING;

-- Drop the temporary function
DROP FUNCTION IF EXISTS map_question_to_domain;

-- Create a view to simplify domain score calculation
CREATE OR REPLACE VIEW ilo_question_domain_view AS
SELECT 
    q.id AS question_id,
    q.question_number,
    q.question_text,
    d.id AS domain_id,
    d.code AS domain_code,
    d.name AS domain_name,
    qd.weight
FROM 
    ilo_questions q
JOIN 
    ilo_question_domains qd ON q.id = qd.question_id
JOIN 
    ilo_domains d ON d.id = qd.domain_id
ORDER BY 
    q.question_number, d.code;

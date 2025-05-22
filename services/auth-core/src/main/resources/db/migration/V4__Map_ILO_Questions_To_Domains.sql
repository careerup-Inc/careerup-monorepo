-- Helper function to map scales to domains
CREATE OR REPLACE FUNCTION map_scale_to_domain(scale_name TEXT)
RETURNS TEXT AS $$
BEGIN
    -- Map Vietnamese ILO scales to new domain codes
    -- This is a simplification; actual mapping should be based on the specific ILO test structure
    IF scale_name ILIKE '%ngôn ngữ%' OR scale_name ILIKE '%văn học%' OR scale_name ILIKE '%đọc viết%' THEN
        RETURN 'LANG';
    ELSIF scale_name ILIKE '%phân tích%' OR scale_name ILIKE '%logic%' OR scale_name ILIKE '%toán học%' OR scale_name ILIKE '%suy luận%' THEN
        RETURN 'LOGIC';
    ELSIF scale_name ILIKE '%thiết kế%' OR scale_name ILIKE '%nghệ thuật%' OR scale_name ILIKE '%màu sắc%' OR scale_name ILIKE '%hình học%' THEN
        RETURN 'DESIGN';
    ELSIF scale_name ILIKE '%con người%' OR scale_name ILIKE '%giao tiếp%' OR scale_name ILIKE '%xã hội%' OR scale_name ILIKE '%nhóm%' THEN
        RETURN 'PEOPLE';
    ELSIF scale_name ILIKE '%thể chất%' OR scale_name ILIKE '%cơ khí%' OR scale_name ILIKE '%kỹ thuật%' OR scale_name ILIKE '%vận động%' THEN
        RETURN 'MECH';
    ELSE
        -- Default to LANG if no match (should be replaced with appropriate logic)
        RETURN 'LANG';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Ensure unique mapping between question and domain
ALTER TABLE ilo_question_domains
    ADD CONSTRAINT uq_question_domain UNIQUE (question_id, domain_id);

-- Map existing questions to domains based on their scale (removed ilo_scales join)
INSERT INTO ilo_question_domains (question_id, domain_id, weight)
SELECT 
    q.id AS question_id,
    d.id AS domain_id,
    1.0 AS weight -- Default weight
FROM 
    ilo_questions q
JOIN 
    ilo_domains d ON d.code = map_scale_to_domain(q.question_text)
ON CONFLICT (question_id, domain_id) DO NOTHING;

-- Drop the temporary function
DROP FUNCTION IF EXISTS map_scale_to_domain;

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

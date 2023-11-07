-- Drop tables that reference others first

DROP TABLE IF EXISTS task_reschedules CASCADE;
DROP TABLE IF EXISTS user_tasks CASCADE;
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS calendar_events CASCADE;
DROP TABLE IF EXISTS evidence CASCADE;
DROP TABLE IF EXISTS user_cases CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS logs CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS app_users CASCADE;
DROP TABLE IF EXISTS cases CASCADE;

-- Now we can safely drop these tables

DROP TABLE IF EXISTS task_types CASCADE;
DROP TABLE IF EXISTS evidence_types CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS role CASCADE;
DROP TABLE IF EXISTS courts CASCADE;
DROP TABLE IF EXISTS case_types CASCADE;

-- Dropping triggers and trigger function
-- Note: These should be dropped before the tables they are related to.

DROP TRIGGER IF EXISTS update_cases_modtime ON cases;
DROP TRIGGER IF EXISTS update_evidence_modtime ON evidence;
DROP TRIGGER IF EXISTS update_app_users_modtime ON app_users;
DROP FUNCTION IF EXISTS update_modified_column();

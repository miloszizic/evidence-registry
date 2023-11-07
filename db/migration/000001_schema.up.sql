CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

--------------------------------------------------------
CREATE TABLE "cases" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  "name" varchar NOT NULL,
  "tags" text[],
  "case_year" int NOT NULL,
  "case_type_id" uuid NOT NULL,
  "case_number" int NOT NULL,
  "case_court_id" uuid NOT NULL
);

CREATE TABLE "case_types" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "name" varchar NOT NULL,
  "description" varchar NOT NULL
);

CREATE TABLE "courts" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "code" int UNIQUE NOT NULL,
  "name" varchar NOT NULL,
  "short_name" varchar NOT NULL
);

CREATE TABLE "user_cases" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "user_id" uuid NOT NULL,
  "case_id" uuid NOT NULL
);

CREATE TABLE "evidence" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "case_id" uuid NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  "app_user_id" uuid NOT NULL,
  "name" varchar NOT NULL,
  "description" varchar,
  "hash" varchar NOT NULL,
  "evidence_type_id" uuid NOT NULL
);

CREATE TABLE "role" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "name" varchar NOT NULL,
  "code" varchar NOT NULL
);

CREATE TABLE "audit_logs" (
  "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "action" VARCHAR(50) NOT NULL,
  "table_name" VARCHAR(255) NOT NULL,
  "record_id" UUID NOT NULL,
  "old_data" TEXT,
  "new_data" TEXT,
  "changed_at" TIMESTAMP NOT NULL DEFAULT (now()),
  "changed_by" UUID
);

CREATE TABLE "app_users" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "username" varchar UNIQUE NOT NULL,
  "email" varchar NOT NULL,
  "password" varchar NOT NULL,
  "role_id" uuid,
  "first_name" varchar,
  "last_name" varchar,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "evidence_types" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "name" varchar NOT NULL
);

CREATE TABLE "calendar_events" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "user_id" uuid NOT NULL,
  "case_id" uuid NOT NULL,
  "event_date" date NOT NULL,
  "notes" varchar,
  "task_id" uuid
);

CREATE TABLE "task_types" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "name" varchar NOT NULL
);

CREATE TABLE "tasks" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "name" varchar NOT NULL,
  "description" varchar,
  "task_type_id" uuid NOT NULL,
  "case_id" uuid
);

CREATE TABLE "user_tasks" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "user_id" uuid NOT NULL,
  "task_id" uuid NOT NULL,
  "assigned_by" uuid NOT NULL,
  "due_date" date NOT NULL,
  "is_completed" boolean DEFAULT false,
  "reschedule_count" int DEFAULT 0
);

CREATE TABLE "task_reschedules" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "user_task_id" uuid NOT NULL,
  "new_due_date" date NOT NULL,
  "reassigned_to" uuid,
  "comment" varchar,
  "rescheduled_by" uuid NOT NULL
);

CREATE TABLE "permissions" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "name" varchar NOT NULL,
  "code" varchar NOT NULL
);

CREATE TABLE "role_permissions" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "role_id" uuid NOT NULL,
  "permission_id" uuid NOT NULL
);

CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "user_id" uuid NOT NULL,
  "refresh_payload_id" uuid NOT NULL,
  "username" varchar NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expires_at" timestamp NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "session_data" (
  "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "key" varchar UNIQUE NOT NULL,
  "value" uuid NOT NULL
);

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "app_users" ("id") ON DELETE CASCADE;

ALTER TABLE "tasks" ADD FOREIGN KEY ("task_type_id") REFERENCES "task_types" ("id");

ALTER TABLE "tasks" ADD FOREIGN KEY ("case_id") REFERENCES "cases" ("id");

ALTER TABLE "user_tasks" ADD FOREIGN KEY ("task_id") REFERENCES "tasks" ("id");

ALTER TABLE "user_tasks" ADD FOREIGN KEY ("user_id") REFERENCES "app_users" ("id") ON DELETE CASCADE;

ALTER TABLE "user_tasks" ADD FOREIGN KEY ("assigned_by") REFERENCES "app_users" ("id") ON DELETE CASCADE;

ALTER TABLE "task_reschedules" ADD FOREIGN KEY ("user_task_id") REFERENCES "user_tasks" ("id");

ALTER TABLE "task_reschedules" ADD FOREIGN KEY ("rescheduled_by") REFERENCES "app_users" ("id") ON DELETE CASCADE;

ALTER TABLE "task_reschedules" ADD FOREIGN KEY ("reassigned_to") REFERENCES "app_users" ("id") ON DELETE CASCADE;

ALTER TABLE "calendar_events" ADD FOREIGN KEY ("user_id") REFERENCES "app_users" ("id") ON DELETE CASCADE;

ALTER TABLE "calendar_events" ADD FOREIGN KEY ("case_id") REFERENCES "cases" ("id");

ALTER TABLE "evidence" ADD FOREIGN KEY ("evidence_type_id") REFERENCES "evidence_types" ("id");

ALTER TABLE "app_users" ADD FOREIGN KEY ("role_id") REFERENCES "role" ("id");

ALTER TABLE "user_cases" ADD FOREIGN KEY ("case_id") REFERENCES "cases" ("id") ON DELETE CASCADE;

ALTER TABLE "cases" ADD FOREIGN KEY ("case_court_id") REFERENCES "courts" ("id");

ALTER TABLE "cases" ADD FOREIGN KEY ("case_type_id") REFERENCES "case_types" ("id");

ALTER TABLE "user_cases" ADD FOREIGN KEY ("user_id") REFERENCES "app_users" ("id") ON DELETE CASCADE;

ALTER TABLE "evidence" ADD FOREIGN KEY ("case_id") REFERENCES "cases" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("role_id") REFERENCES "role" ("id") ON DELETE CASCADE;

ALTER TABLE "calendar_events" ADD FOREIGN KEY ("task_id") REFERENCES "tasks" ("id");



-----------------------------------------------------------------------



-- Creating trigger function
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$ language 'plpgsql';

---- Creating trigger function for audit logs

--SET myapp.current_user = 'default';


---- Cases
CREATE OR REPLACE FUNCTION audit_cases_changes()
RETURNS TRIGGER AS $$
DECLARE
    current_user_uuid UUID;
BEGIN
   -- Fetch the current_user from the temporary table
   SELECT value::uuid INTO current_user_uuid FROM session_data WHERE key = 'current_user';

   IF TG_OP = 'DELETE' THEN
      INSERT INTO audit_logs(action, table_name, record_id, old_data, changed_by)
      VALUES('DELETE', 'cases', OLD.id, row_to_json(OLD)::text, current_user_uuid);
      RETURN OLD;
   ELSIF TG_OP = 'UPDATE' THEN
      INSERT INTO audit_logs(action, table_name, record_id, old_data, new_data, changed_by)
      VALUES('UPDATE', 'cases', NEW.id, row_to_json(OLD)::text, row_to_json(NEW)::text, current_user_uuid);
      RETURN NEW;
   ELSIF TG_OP = 'INSERT' THEN
      INSERT INTO audit_logs(action, table_name, record_id, new_data, changed_by)
      VALUES('INSERT', 'cases', NEW.id, row_to_json(NEW)::text, current_user_uuid);
      RETURN NEW;
   END IF;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER audit_cases_trigger
AFTER INSERT OR UPDATE OR DELETE ON cases
FOR EACH ROW EXECUTE FUNCTION audit_cases_changes();

-- Evidence
CREATE OR REPLACE FUNCTION audit_evidence_changes()
RETURNS TRIGGER AS $$
DECLARE
    current_user_uuid UUID;
BEGIN
   -- Fetch the current_user from the session_data table
   SELECT value::uuid INTO current_user_uuid FROM session_data WHERE key = 'current_user';

   IF TG_OP = 'DELETE' THEN
      INSERT INTO audit_logs(action, table_name, record_id, old_data, changed_by)
      VALUES('DELETE', 'evidence', OLD.id, row_to_json(OLD)::text, current_user_uuid);
      RETURN OLD;
   ELSIF TG_OP = 'UPDATE' THEN
      INSERT INTO audit_logs(action, table_name, record_id, old_data, new_data, changed_by)
      VALUES('UPDATE', 'evidence', NEW.id, row_to_json(OLD)::text, row_to_json(NEW)::text, current_user_uuid);
      RETURN NEW;
   ELSIF TG_OP = 'INSERT' THEN
      INSERT INTO audit_logs(action, table_name, record_id, new_data, changed_by)
      VALUES('INSERT', 'evidence', NEW.id, row_to_json(NEW)::text, current_user_uuid);
      RETURN NEW;
   END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_evidence_trigger
AFTER INSERT OR UPDATE OR DELETE ON evidence
FOR EACH ROW EXECUTE FUNCTION audit_evidence_changes();


------------------------------------------------
-- Adding some data in tables for testing purposes

INSERT INTO courts (code, name, short_name) VALUES
   (42, 'APELACIONI SUD CG', 'ASCG'),
   (14, 'OSNOVNI SUD U BARU', 'OSBR'),
   (23, 'OSNOVNI SUD U BERANAMA', 'OSBE'),
   (20, 'OSNOVNI SUD U BIJELOM POLJU', 'OSBP'),
   (34, 'OSNOVNI SUD U CETINJU', 'OSCT'),
   (35, 'OSNOVNI SUD U DANILOVGRADU', 'OSDA'),
   (33, 'OSNOVNI SUD U HERCEG NOVOM', 'OSHN'),
   (38, 'OSNOVNI SUD U KOLAŠINU', 'OSKO'),
   (15, 'OSNOVNI SUD U KOTORU', 'OSKT'),
   (36, 'OSNOVNI SUD U NIKŠIĆU', 'OSNK'),
   (40, 'OSNOVNI SUD U PLAVU', 'OSPL'),
   (39, 'OSNOVNI SUD U PLJEVLJIMA', 'OSPV'),
   (9, 'OSNOVNI SUD U PODGORICI', 'OSPG'),
   (41, 'OSNOVNI SUD U ROŽAJAMA', 'OSRO'),
   (11, 'OSNOVNI SUD U ULCINJU', 'OSUL'),
   (37, 'OSNOVNI SUD U ŽABLJAKU', 'OSŽA'),
   (18, 'PRIVREDNI SUD BIJELO POLJE', 'PSBP'),
   (4, 'PRIVREDNI SUD CRNE GORE', 'PSCG'),
   (46, 'SUD ZA PREKRŠAJE BIJELO POLJE', 'SZPBP'),
   (47, 'SUD ZA PREKRŠAJE BUDVA', 'SZPBU'),
   (143, 'SUD ZA PREKRŠAJE HERCEG NOVI', 'SZPHN'),
   (8, 'SUD ZA PREKRŠAJE PODGORICA', 'SZPPG'),
   (43, 'UPRAVNI SUD CG', 'USCG'),
   (21, 'VIŠI SUD U BIJELOM POLJU', 'VSBP'),
   (12, 'VIŠI SUD U PODGORICI', 'VSPG'),
   (3, 'VIŠI SUD ZA PREKRŠAJE CG', 'VSZPCG'),
   (13, 'VRHOVNI SUD CG', 'VSCG');



INSERT INTO case_types (name, description) VALUES
   ('KM', 'KRIVIČNI POSTUPAK PREMA MALOLJETNICIMA'),
   ('MAL', 'PARNIČNI PREDMETI MALE VRIJEDNOSTI'),
   ('K', 'PRVOSTEPENI KRIVIČNI PREDMETI'),
   ('KS', 'KRIVIČNI PREDMETI - SPECIJALNI'),
   ('KV', 'KRIVIČNO VIJEĆE VAN GLAVNOG PRETRESA'),
   ('P', 'PARNIČNI PREDMETI');

-- Adding permissions
INSERT INTO permissions (name, code) VALUES
   ('view_case', 'VCASE'),
   ('create_case', 'CCASE'),
   ('edit_case', 'ECASE'),
   ('delete_case', 'DCASE'),
   ('view_evidence', 'VEVID'),
   ('create_evidence', 'CEVID'),
   ('edit_evidence', 'EEVID'),
   ('delete_evidence', 'DEVID'),
   ('view_user', 'VUSER'),
   ('create_user', 'CUSER'),
   ('delete_user', 'DUSER'),
   ('create_role', 'CROLE'),
   ('view_role', 'VROLE'),
   ('edit_role', 'EROLE'),
   ('delete_role', 'DROLE');

-- Adding some default roles
INSERT INTO role (name, code) VALUES
   ('admin', 'ADMIN'),
   ('editor', 'EDIT'),
   ('viewer', 'VIEW');

-- Adding some initial evidence types
INSERT INTO evidence_types (name) VALUES
  ('Initial Evidence'),
  ('Processed Evidence'),
  ('New Evidence');


-- Giving 'admin' all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT role.id, permissions.id
FROM role, permissions
WHERE role.code = 'ADMIN' AND permissions.code IN ('VCASE', 'CCASE', 'ECASE', 'DCASE', 'VEVID', 'CEVID', 'EEVID', 'DEVID', 'VUSER', 'CUSER', 'DUSER', 'CROLE','VROLE', 'EROLE', 'DROLE');

-- Giving 'editor' only 'view' and 'edit' permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT role.id, permissions.id
FROM role, permissions
WHERE role.code = 'EDIT' AND permissions.code IN ('VCASE', 'ECASE', 'VEVID', 'EEVID', 'VUSER', 'CUSER', 'VROLE', 'EROLE');

-- Giving 'viewer' only 'view' permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT role.id, permissions.id
FROM role, permissions
WHERE role.code = 'VIEW' AND permissions.code IN ('VCASE', 'VEVID', 'VUSER', 'VROLE');


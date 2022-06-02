CREATE TABLE IF NOT EXISTS "users" (
  	"id" SERIAL,
	"username"	VARCHAR(255) NOT NULL,
	"password"	VARCHAR(255) NOT NULL,
	"role"		VARCHAR(255) NOT NULL DEFAULT 'user',
	PRIMARY KEY("id")
);
CREATE TABLE IF NOT EXISTS "cases" (
	"id" SERIAL,
	"name"	VARCHAR(255) NOT NULL,
	"tags"	text[] ,
	PRIMARY KEY("id")
);
CREATE TABLE IF NOT EXISTS "user_cases" (
	"user_id"	integer,
	"case_id"	integer,
	PRIMARY KEY("user_id","case_id"),
	CONSTRAINT "fk_user_cases_case" FOREIGN KEY("case_id") REFERENCES "cases"("id"),
	CONSTRAINT "fk_user_cases_user" FOREIGN KEY("user_id") REFERENCES "users"("id")
);
CREATE TABLE IF NOT EXISTS "evidences" (
	"id" SERIAL,
	"case_id"	integer,
	"name"	VARCHAR(255) NOT NULL,
	"hash"	VARCHAR(255) NOT NULL,
	PRIMARY KEY("id"),
	CONSTRAINT "fk_cases_evidence" FOREIGN KEY("case_id") REFERENCES "cases"("id")
);

CREATE TABLE IF NOT EXISTS "comments" (
	"id" SERIAL,
	"evidence_id"	integer,
	"content"	VARCHAR(255) NOT NULL,
	PRIMARY KEY("id"),
--    CONSTRAINT "fk_comments_user" FOREIGN KEY("user_id") REFERENCES "users"("id"),
	CONSTRAINT "fk_comments_evidence" FOREIGN KEY("evidence_id") REFERENCES "evidences"("id")
);

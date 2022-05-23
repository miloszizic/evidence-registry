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
	CONSTRAINT "fk_cases_objects" FOREIGN KEY("case_id") REFERENCES "cases"("id")
);

CREATE TABLE IF NOT EXISTS "comments" (
	"id" SERIAL,
	"evidence_id"	integer,
	"content"	VARCHAR(255) NOT NULL,
	PRIMARY KEY("id"),
--    CONSTRAINT "fk_comments_user" FOREIGN KEY("user_id") REFERENCES "users"("id"),
	CONSTRAINT "fk_comments_evidence" FOREIGN KEY("evidence_id") REFERENCES "evidences"("id")
);
--CREATE TABLE IF NOT EXISTS "cases_tags" (
--	"case_id"	integer,
--	"tag"	VARCHAR(255) NOT NULL,
--	PRIMARY KEY("case_id","tag"),
--	CONSTRAINT "fk_cases_tags_case" FOREIGN KEY("case_id") REFERENCES "cases"("id")
--)
--CREATE TABLE IF NOT EXISTS "evidence_comments" (
--	"evidence_id"	integer,
--	"comment_id"	integer,
--	PRIMARY KEY("evidence_id","comment_id"),
--	CONSTRAINT "fk_evidence_comments_comment" FOREIGN KEY("comment_id") REFERENCES "comments"("id"),
--	CONSTRAINT "fk_evidence_comments_evidence" FOREIGN KEY("evidence_id") REFERENCES "evidences"("id")
--);


--
--CREATE SCHEMA IF NOT EXISTS evidences;
---- ************************************** evidences.roles
--CREATE TABLE IF NOT EXISTS evidences.roles
--(
-- "id"   int NOT NULL,
-- name text NOT NULL,
-- CONSTRAINT PK_75 PRIMARY KEY ( "id" )
--);
---- ************************************** evidences.registers
--CREATE TABLE IF NOT EXISTS evidences.registers
--(
-- "id"   int NOT NULL,
-- name text NOT NULL,
-- CONSTRAINT PK_56 PRIMARY KEY ( "id" )
--);
---- ************************************** evidences.labels
--CREATE TABLE IF NOT EXISTS evidences.labels
--(
-- "id"   int NOT NULL,
-- name text NOT NULL,
-- CONSTRAINT PK_42 PRIMARY KEY ( "id" )
--);
---- ************************************** evidences.courts
--CREATE TABLE IF NOT EXISTS evidences.courts
--(
-- "id"   int NOT NULL,
-- name text NOT NULL,
-- CONSTRAINT PK_35 PRIMARY KEY ( "id" )
--);
---- ************************************** evidences.cities
--CREATE TABLE IF NOT EXISTS evidences.cities
--(
-- "id"   int NOT NULL,
-- name text NOT NULL,
-- CONSTRAINT PK_26 PRIMARY KEY ( "id" )
--);
---- ************************************** evidences.services
--CREATE TABLE IF NOT EXISTS evidences.users
--(
-- "id"       int NOT NULL,
-- username text NOT NULL,
-- role_id  int NOT NULL,
-- password text NOT NULL,
-- CONSTRAINT PK_51 PRIMARY KEY ( "id" ),
-- CONSTRAINT FK_77 FOREIGN KEY ( role_id ) REFERENCES evidences.roles ( "id" )
--);
--
--CREATE INDEX FK_79 ON evidences.users
--(
-- role_id
--);
---- ************************************** evidences.cases
--CREATE TABLE IF NOT EXISTS evidences.cases
--(
-- "id"          text NOT NULL,
-- created_at  date NOT NULL,
-- user_id     int NOT NULL,
-- reg_id      int NOT NULL,
-- court_id    int NOT NULL,
-- deleted_at  date NOT NULL,
-- case_year   int NOT NULL,
-- case_numb   int NOT NULL,
-- modified_at date NOT NULL,
-- CONSTRAINT PK_6 PRIMARY KEY ( "id" ),
-- CONSTRAINT FK_37 FOREIGN KEY ( court_id ) REFERENCES evidences.courts ( "id" ),
-- CONSTRAINT FK_58 FOREIGN KEY ( reg_id ) REFERENCES evidences.registers ( "id" ),
-- CONSTRAINT FK_87 FOREIGN KEY ( user_id ) REFERENCES evidences.users ( "id" )
--);
--CREATE INDEX FK_39 ON evidences.cases
--(
-- court_id
--);
--CREATE INDEX FK_60 ON evidences.cases
--(
-- reg_id
--);
--
--CREATE INDEX FK_89 ON evidences.cases
--(
-- user_id
--);
---- ************************************** evidences.evidences
--CREATE TABLE IF NOT EXISTS evidences.evidences
--(
-- "id"          int NOT NULL,
-- case_id     text NOT NULL,
-- city_id     int NOT NULL,
-- name        text NOT NULL,
-- hash        text NOT NULL,
-- anotation   text NOT NULL,
-- object_ref  text NOT NULL,
-- created_at  date NOT NULL,
-- deleted_at  date NOT NULL,
-- modified_at date NOT NULL,
-- CONSTRAINT PK_21 PRIMARY KEY ( "id" ),
-- CONSTRAINT FK_17 FOREIGN KEY ( case_id ) REFERENCES evidences.cases ( "id" ),
-- CONSTRAINT FK_29 FOREIGN KEY ( city_id ) REFERENCES evidences.cities ( "id" )
--);
--
--CREATE INDEX FK_19 ON evidences.evidences
--(
-- case_id
--);
--
--CREATE INDEX FK_31 ON evidences.evidences
--(
-- city_id
--);
--
---- ************************************** evidences.case_lbls
--
--CREATE TABLE IF NOT EXISTS evidences.case_lbls
--(
-- lbls_id int NOT NULL,
-- case_id text NOT NULL,
-- CONSTRAINT FK_64 FOREIGN KEY ( lbls_id ) REFERENCES evidences.labels ( "id" ),
-- CONSTRAINT FK_70 FOREIGN KEY ( case_id ) REFERENCES evidences.cases ( "id" )
--);
--
--CREATE INDEX FK_66 ON evidences.case_lbls
--(
-- lbls_id
--);
--
--CREATE INDEX FK_72 ON evidences.case_lbls
--(
-- case_id
--);

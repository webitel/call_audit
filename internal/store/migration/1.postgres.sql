CREATE SCHEMA call_audit;


-- call_audit.call_questionnaire_rule definition

-- Drop table

-- DROP TABLE call_audit.call_questionnaire_rule;

CREATE TABLE call_audit.call_questionnaire_rule (
	id serial4 NOT NULL,
	domain_id int8 NOT NULL,
	created_at timestamptz DEFAULT now() NOT NULL,
	created_by int8 NOT NULL,
	updated_at timestamptz DEFAULT now() NOT NULL,
	updated_by int8 NOT NULL,
	"name" varchar NOT NULL,
	language_profile int4 NOT NULL,
	description varchar NULL,
	enabled bool NULL,
	сognitive_profile int4 NOT NULL,
	"from" timestamptz NOT NULL,
	"to" timestamptz NULL,
	call_direction varchar NULL,
	min_call_duration int4 NULL,
	variable varchar NULL,
	default_promt varchar NULL,
	save_explanation bool NULL,
	last_stored_at timestamptz NULL,
	CONSTRAINT call_questionnaire_rule_pk PRIMARY KEY (id),
	CONSTRAINT call_questionnaire_rule_cognitive_profile_services_fk FOREIGN KEY (сognitive_profile) REFERENCES "storage".cognitive_profile_services(id) ON DELETE SET NULL,
	CONSTRAINT call_questionnaire_rule_language_profiles_fk FOREIGN KEY (language_profile) REFERENCES "storage".language_profiles(id),
	CONSTRAINT call_questionnaire_rule_wbt_domain_fk FOREIGN KEY (domain_id) REFERENCES directory.wbt_domain(dc),
	CONSTRAINT call_questionnaire_rule_wbt_user_fk FOREIGN KEY (created_by) REFERENCES directory.wbt_user(id),
	CONSTRAINT call_questionnaire_rule_wbt_user_fk_1 FOREIGN KEY (updated_by) REFERENCES directory.wbt_user(id)
);

-- Permissions

ALTER TABLE call_audit.call_questionnaire_rule OWNER TO opensips;
GRANT ALL ON TABLE call_audit.call_questionnaire_rule TO opensips;


-- "storage".language_profiles definition

-- Drop table

-- DROP TABLE "storage".language_profiles;

CREATE TABLE "storage".language_profiles (
	id int4 NOT NULL,
	domain_id int4 NOT NULL,
	created_at timestamptz DEFAULT now() NULL,
	created_by int8 NOT NULL,
	updated_at timestamptz DEFAULT now() NULL,
	updated_by int8 NOT NULL,
	"name" varchar NOT NULL,
	"token" varchar NULL,
	"type" int4 NOT NULL,
	CONSTRAINT language_models_pk PRIMARY KEY (id)
);

-- Permissions

ALTER TABLE "storage".language_profiles OWNER TO opensips;
GRANT ALL ON TABLE "storage".language_profiles TO opensips;
GRANT SELECT ON TABLE "storage".language_profiles TO grafana;


-- call_audit.jobs definition

-- Drop table

-- DROP TABLE call_audit.jobs;

CREATE UNLOGGED TABLE call_audit.jobs (
	id serial4 NOT NULL,
	rule_id int4 NULL,
	"type" int4 NULL,
	params jsonb NULL,
	state int4 DEFAULT 0 NULL,
	call_storead_at timestamptz NULL
);

-- Permissions

ALTER TABLE call_audit.jobs OWNER TO opensips;
GRANT ALL ON TABLE call_audit.jobs TO opensips;
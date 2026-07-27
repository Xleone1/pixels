--liquibase formatted sql
--changeset pixels:pixels-progression-0010-audit-state
alter table progression_audit
    add column before_state jsonb not null default '{}'::jsonb,
    add column after_state jsonb not null default '{}'::jsonb;
--rollback alter table progression_audit drop column if exists after_state, drop column if exists before_state;

package main

import (
	"fmt"
	"strconv"
	"strings"
)

// renderSQL produces deterministic reversible Liquibase SQL.
func renderSQL(value manifest) []byte {
	var builder strings.Builder
	builder.WriteString("--liquibase formatted sql\n\n")
	builder.WriteString("--changeset pixels:furniture-seed-posture-metadata-0046 context:development\n")
	builder.WriteString("create temporary table posture_review(id bigint primary key,name text not null,status text not null,confidence text not null,source text not null,reason text not null,previous_allow_sit boolean not null,previous_allow_lay boolean not null) on commit drop;\n")
	builder.WriteString("insert into posture_review values\n")
	for index, item := range value.Reviews {
		builder.WriteString(fmt.Sprintf(
			"(%d,'%s','%s','%s','%s','%s',%s,%s)%s\n",
			item.ID, quoteSQL(item.Name), item.Status, item.Confidence, item.Source,
			quoteSQL(item.Reason), strconv.FormatBool(item.PreviousAllowSit),
			strconv.FormatBool(item.PreviousAllowLay), separator(index, len(value.Reviews)),
		))
	}
	builder.WriteString(";\n")
	builder.WriteString(renderValidation())
	builder.WriteString(renderReviewedUpdate(value.AuditID))
	builder.WriteString(renderDerivedUpdate(value.AuditID))
	builder.WriteString(renderRollback(value))
	return []byte(builder.String())
}

// renderValidation produces guards against stale or manually-owned records.
func renderValidation() string {
	return "create temporary table posture_guard(" +
		"definitions_match boolean constraint posture_review_definitions_must_match check(definitions_match)," +
		"manual_metadata_free boolean constraint posture_review_must_not_override_manual_metadata check(manual_metadata_free)) on commit drop;\n" +
		"insert into posture_guard select " +
		"(select count(*) from posture_review review join furniture_definitions definition on definition.id=review.id and definition.name=review.name and definition.deleted_at is null)=(select count(*) from posture_review)," +
		"not exists(select 1 from posture_review review join furniture_definitions definition on definition.id=review.id where jsonb_exists(definition.metadata,'slots') or coalesce((definition.metadata->'posture'->>'manual_override')::boolean,false));\n"
}

// renderReviewedUpdate produces the evidence-backed posture update.
func renderReviewedUpdate(auditID string) string {
	return "update furniture_definitions definition set " +
		"allow_sit=definition.allow_sit or review.status='sit'," +
		"allow_lay=definition.allow_lay or review.status='lay'," +
		"metadata=jsonb_set(definition.metadata,'{posture}',jsonb_build_object(" +
		"'audit_id','" + quoteSQL(auditID) + "','version',1,'source',review.source," +
		"'status',review.status,'confidence',review.confidence,'reason',review.reason," +
		"'strategy',case when review.status='none' then 'none' else 'derived_footprint' end," +
		"'manual_override',true),true)," +
		"updated_at=now(),version=definition.version+1 " +
		"from posture_review review where definition.id=review.id;\n"
}

// renderDerivedUpdate annotates unreviewed definitions without changing flags.
func renderDerivedUpdate(auditID string) string {
	return "update furniture_definitions definition set " +
		"metadata=jsonb_set(definition.metadata,'{posture}',jsonb_build_object(" +
		"'audit_id','" + quoteSQL(auditID) + "','version',1," +
		"'source',case when jsonb_exists(definition.metadata,'slots') then 'explicit_slots' else 'definition_flags' end," +
		"'status',case when definition.allow_sit and definition.allow_lay then 'sit_lay' when definition.allow_lay then 'lay' when definition.allow_sit then 'sit' else 'none' end," +
		"'confidence',case when jsonb_exists(definition.metadata,'slots') then 'high' when definition.allow_sit or definition.allow_lay then 'medium' else 'unclassified' end," +
		"'strategy',case when jsonb_exists(definition.metadata,'slots') then 'explicit_slots' when definition.allow_sit or definition.allow_lay then 'derived_footprint' else 'none' end," +
		"'manual_override',false),true)," +
		"updated_at=now(),version=definition.version+1 " +
		"where definition.deleted_at is null and not jsonb_exists(definition.metadata,'posture');\n"
}

// renderRollback restores flags and removes only this audit's metadata.
func renderRollback(value manifest) string {
	var sitIDs []string
	var layIDs []string
	for _, item := range value.Reviews {
		if item.Status == "sit" {
			sitIDs = append(sitIDs, strconv.FormatInt(item.ID, 10))
		}
		if item.Status == "lay" {
			layIDs = append(layIDs, strconv.FormatInt(item.ID, 10))
		}
	}
	return "--rollback update furniture_definitions set allow_sit=false where id in (" +
		strings.Join(sitIDs, ",") + "); update furniture_definitions set allow_lay=false where id in (" +
		strings.Join(layIDs, ",") + "); update furniture_definitions set metadata=metadata-'posture' where metadata->'posture'->>'audit_id'='" +
		quoteSQL(value.AuditID) + "';\n"
}

// quoteSQL escapes one PostgreSQL string literal body.
func quoteSQL(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

// separator returns a comma except after the final SQL row.
func separator(index int, total int) string {
	if index == total-1 {
		return ""
	}
	return ","
}

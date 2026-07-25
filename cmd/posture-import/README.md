# Furniture posture import

This command turns a reviewed posture manifest into a reversible Liquibase
seed. It never edits the database directly.

```bash
go run ./cmd/posture-import \
  -reviews internal/realm/furniture/database/seed/source/posture_reviews_20260725.json \
  -output internal/realm/furniture/database/seed/development/0046_posture_metadata.sql
```

The import uses this evidence order:

1. Existing explicit `metadata.slots` and manual overrides.
2. Trusted `FurnitureData` posture flags.
3. Visual review of the public furniture icon.
4. Existing definition flags with runtime footprint-derived slots.

`.nitro` assets validate footprint and visual geometry, but do not provide
reliable sit or lay semantics. The generated SQL therefore preserves explicit
slots and refuses to overwrite earlier manual metadata.

Every reviewed decision records its source, confidence, reason, strategy, and
audit identifier in `metadata.posture`. Unreviewed definitions are annotated
without changing their capability flags. The generated rollback restores the
reviewed flags and removes only metadata belonging to the matching audit.

# 2025-11-22 – Runtime Layering: Metadata migration/validation tools

## Summary
- Stopped auto-using legacy metadata buckets: `bootstrapMetadataStore` now fails fast when legacy rows are present so runtime stays `metadata_v2`-only.
- Added `onemount --metadata-validate` / `--metadata-migrate-legacy` maintenance path that validates `metadata_v2` or migrates legacy buckets into v2 and drops the legacy bucket.
- Introduced reusable Bolt helpers (`ValidateMetadataBucket`, `MigrateLegacyMetadata`) plus tests covering the new migration gating.

## Testing
- `go test ./internal/fs -run "Metadata" -count=1`
- `go test ./cmd/onemount -run Test -count=1`

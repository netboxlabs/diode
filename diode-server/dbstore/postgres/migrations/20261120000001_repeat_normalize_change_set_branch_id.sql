-- +goose Up
-- +goose StatementBegin
-- Same reason as in 20261118000001: the UPDATE scans every change_set.
SET LOCAL statement_timeout = '30min';
-- +goose StatementEnd

-- +goose StatementBegin
-- Repeat of 20261118000001. That migration ran before a rolling update
-- replaced the old reconciler pods, which kept writing branch_id as
-- "<name> (<schema_id>)" until they stopped. Rewrite those rows to the bare
-- schema id. The match is the same, so rows that are already bare are not
-- changed.
UPDATE change_sets
SET branch_id = substring(branch_id FROM ' \(([a-z0-9]{8})\)$')
WHERE branch_id ~ ' \([a-z0-9]{8}\)$';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- The branch name is not recoverable from the schema id; nothing to undo.
SELECT 1;
-- +goose StatementEnd

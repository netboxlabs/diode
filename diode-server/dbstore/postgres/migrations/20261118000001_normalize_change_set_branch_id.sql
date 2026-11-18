-- +goose Up
-- +goose StatementBegin
-- Successful plans stored change_sets.branch_id as "<name> (<schema_id>)",
-- while failed-plan placeholders stored the bare schema id. Branch filters
-- and dedup compare the column by equality, so one branch had two keys.
-- The reconciler now stores only the schema id; rewrite the older rows to
-- match.
--
-- A NetBox branching schema id is always 8 characters of [a-z0-9]. Only a
-- value that ends in exactly that, in parentheses after a space, is
-- rewritten. A bare schema id has no space or parentheses, so it never
-- matches and the migration is safe to run again.
UPDATE change_sets
SET branch_id = substring(branch_id FROM ' \(([a-z0-9]{8})\)$')
WHERE branch_id ~ ' \([a-z0-9]{8}\)$';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- The branch name is not recoverable from the schema id; nothing to undo.
SELECT 1;
-- +goose StatementEnd

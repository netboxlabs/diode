-- +goose Up

-- Add trigger to change_sets table
-- +goose StatementBegin
CREATE TRIGGER update_change_sets_updated_at
    BEFORE UPDATE ON change_sets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- Add trigger to changes table
-- +goose StatementBegin
CREATE TRIGGER update_changes_updated_at
    BEFORE UPDATE ON changes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down

-- Drop trigger from changes table
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_changes_updated_at ON changes;
-- +goose StatementEnd

-- Drop trigger from change_sets table
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_change_sets_updated_at ON change_sets;
-- +goose StatementEnd
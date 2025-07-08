# Migrations

## Creating new migrations
Add new migrations using the following command, changing the name parameter to describe what the new migration will do.
```bash
make create-migration name=your_migration_name
```

See the [goose documentation](https://github.com/pressly/goose#sql-migrations) for details on how to write migrations.
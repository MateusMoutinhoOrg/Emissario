package reindex

import (
	api "github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	databasedeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/databasedeps"
)

// Database runs the store's Repair over every table of database that indexes
// a key field, its own or one of a table nested under its records, which
// Repair recurses into: every live record gets back every index entry it is
// missing, and what a write stopped in the middle left behind is deleted. A
// table with no key field has no index to rebuild and is left alone, since
// Repair reads every id the table ever gave. Only the methods_custom.go of a
// database's package sees its store database, so each one hands it here.
func Database(sandbox *api.Sandbox, database databasedeps.Database) error {
	if database.Collection == nil {
		return sandbox.Deps.StdDeps.Errorf("a database was never built: Databases.New refused its Props")
	}
	for _, schema := range database.Props.Schemas {
		if !indexed(schema.Fields) {
			continue
		}
		collection, err := sandbox.Deps.OpinionatedAgnosDatabase.Collection(database, schema.Name)
		if err != nil {
			return err
		}
		if failure := collection.Repair(); failure != nil {
			return sandbox.Deps.StdDeps.Errorf("%s: %s", schema.Name, sandbox.Deps.OpinionatedAgnosDatabase.Fail(failure).Error())
		}
	}
	return nil
}

// indexed tells whether fields declare a key field, at any depth.
func indexed(fields []databasedeps.Field) bool {
	for _, field := range fields {
		if field.Type == databasedeps.Key || field.Type == databasedeps.Nested && indexed(field.Fields) {
			return true
		}
	}
	return false
}

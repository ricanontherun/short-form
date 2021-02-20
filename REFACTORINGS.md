# Command Structure

1. CQRS
   a. Handler::WriteNote -> WriteNoteCommand
   b. Handler::SearchNotes -> SearchNotesQuery
2. External dependencies are bleeding into module code.
   a. Resolve cli.context to command specific objects/DTOs to decouple the module from the CLI lib.
   b. testing will be easier, no longer requiring the (awkward) construction of a cli.Context.
3. Builder pattern as implemented is clunky.
4. Research SQL querybuilders (or, perhaps an ORM) for use within the existing Repository structure.
5. Not a refactoring, but explore how panic()s would be handled if raised in current arch.
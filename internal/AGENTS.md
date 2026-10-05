# Go implementation guidance

Read the relevant contract in ../docs/specifications/ and component boundary in
../docs/architecture/ before changing controller, runtime or persisted schemas.
Preserve absent-field legacy serialization and deterministic replay when adding
immutable inputs. Role contracts remain executable Go behavior; Markdown
workflows cannot change capabilities, identities or effect authority.

Use ../CONTRIBUTING.md for toolchains and requested checks. Changes to durable
records need explicit compatibility reasoning in the diff and documentation.

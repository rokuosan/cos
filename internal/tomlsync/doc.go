// Package tomlsync provides lightweight TOML document parsing, synthesis, and
// writing helpers for config sync workflows.
//
// The parser intentionally supports only the shape needed by this repository:
// root key/value blocks, table headers, literal and basic quoted table keys,
// trailing comments on headers and values, and bracket-delimited multiline
// collection values whose depth must not be confused with table headers. It
// preserves block structure and text content for those supported shapes, but it
// does normalize line endings to LF and emits a trailing newline per block.
//
// It is not a full TOML implementation and should not be treated as a general
// purpose parser. In particular, it is not intended to support arbitrary
// multiline string forms or broad TOML validation. If future sync features need
// edits inside arbitrary values, the implementation should be replaced with a
// TOML library rather than extended indefinitely.
package tomlsync

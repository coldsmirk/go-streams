# Agent instructions

## Testing

- Write assertions with `testify`: `assert` for a check the test should survive,
  `require` for one it cannot continue past. Do not hand-roll `if got != want {
  t.Errorf(...) }`.
- Map the two by what the native code did: a former `t.Errorf` becomes an
  `assert` call, a former `t.Fatalf` or `t.Fatal` becomes a `require` call.
  Setup that later assertions depend on — opening a file, decoding a fixture,
  an `err` that makes the rest meaningless — is always `require`.
- A length or presence check that guards an index is `require`, and the fields
  it guards are `assert`: `require.Len(t, got, 2)` then `assert.Equal(t, "ada",
  got[0]["name"])`. A single `if len(got) != 2 || got[0][…] != …` guard was the
  native way to get that abort without `require`, so reading it as one
  non-fatal check understates what it did.
- Compare an expected empty slice with `assert.Empty`, never `assert.Equal`
  against `nil`. `slices.Equal(nil, []T{})` is true but `assert.Equal` compares
  with `reflect.DeepEqual`, which separates a nil slice from an empty one, so
  `assert.Equal` would tighten the assertion without saying so.
- End a polling loop that has already decided the outcome with `assert.Fail`,
  not with a fresh assertion on the value it was polling. Re-reading a
  goroutine count or a deadline after the loop gives the failure a second
  chance to pass.
- Use the typed helpers where they read better than a boolean: `assert.Len`,
  `assert.NoError`, `assert.ErrorIs`, `assert.Panics`, `assert.NotPanics`,
  `assert.True`. `assert.Equal` takes `(t, want, got)` in that order.
- Give every assertion a message when the call site alone does not say which
  case failed — the `f` variants (`assert.Equalf`) take a format string. A
  table-driven case that already names itself through `t.Run` does not need one.
- Keep `Example` functions and benchmarks in plain Go. An example is compiled
  documentation checked against its `// Output:` block, and a benchmark has
  nothing to assert.
- Every operation that forwards elements gets an early-termination test. The
  `iter` contract panics if `yield` is called after it returns false, so
  stopping correctly is a correctness requirement, not a nicety.

## Documentation

- State a rule once. Package-wide facts belong in `doc.go` and the README's
  "Semantics worth knowing"; a doc comment states what its own symbol does and
  notes where that symbol is an exception to the rule.
- Document encounter order at the source that fixes it, not on the type that
  carries it.
- Keep `Stream` and `Stream2` symmetric. A method that exists on both is
  documented with the same wording on both, and a fix to one is applied to the
  other in the same change.

## Linting

- `task check` runs the formatter, `go vet`, the linter and the tests.
- The linter version is pinned in CI. Raise it deliberately, not incidentally.
- Prefer a rule exception in `.golangci.yml`, carrying the reason, over a
  `//nolint` comment in the source.

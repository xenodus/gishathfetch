---
name: fix-go-version
description: Check, upgrade, and validate the repository's Go version across configs.
---

Check and update this repository's Go version end-to-end.

## Goal

Keep Go version declarations consistent across all relevant files (module, CI, Docker, tooling) and verify the project still builds/tests correctly after the upgrade.

## Execution Rules

1. Prefer the smallest safe update by default:
   - If no target is provided, update to the latest patch in the current minor series.
   - If explicitly asked, upgrade to the requested Go version.
2. Keep all Go version references aligned across files.
3. Avoid unrelated refactors.
4. Do not edit vendored dependencies except through official tooling commands.
5. Call out risk clearly if a minor/major Go upgrade is performed.

## Steps

1. Detect current Go version usage:
   - Read `go` and `toolchain` directives in `go.mod`.
   - Locate pinned Go versions in CI/workflow files, Dockerfiles, and build scripts.
   - Confirm current local/runtime Go version with `go version` when available.

2. Choose target version:
   - Use the user-provided target version if given.
   - Otherwise determine the recommended patch target and explain why.

3. Implement version updates:
   - Update `go.mod` `go` and/or `toolchain` directives as needed.
   - Update CI setup (for example `actions/setup-go` inputs or `go-version-file` usage).
   - Update Docker/base image tags and any other pinned Go runtime references.
   - Run module maintenance commands required by the ecosystem (for example `go mod tidy`; if vendoring is used, run `go mod vendor` as needed).

4. Validate:
   - Run formatting/lint/build/test checks appropriate for the repo.
   - At minimum run targeted build/test commands for Go components touched.
   - Confirm no stale references to old Go versions remain in maintained files.

5. Prepare delivery:
   - Summarize all old -> new version updates and files changed.
   - Include commands run and outcomes.
   - If writes are allowed, commit and push with a clear message.
   - If PR tooling is available, open/update PR details.

## Final Response Format

Return:

1. `Version Plan`:
   - Current version(s), target version, and upgrade scope (patch/minor/major).
2. `Changes Made`:
   - File-by-file version updates.
3. `Validation`:
   - Commands run and pass/fail outcomes.
4. `Risk Notes`:
   - Any compatibility concerns or deferred follow-up work.

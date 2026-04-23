# Skill: fix-cves

## Purpose

Resolve known vulnerabilities with the smallest safe dependency updates while preserving runtime behavior.

## Inputs

- CVE IDs, advisory links, or Dependabot PRs (if provided)
- Affected manifests and lock/vendor files
- Current test/build commands from repository docs

## Guardrails

1. Prefer patch or minor upgrades before major upgrades.
2. Keep fixes focused on vulnerable dependencies; avoid unrelated refactors.
3. Do not downgrade dependencies to suppress alerts.
4. Apply repository priorities when trade-offs appear:
   - security > correctness > data integrity > performance > clean code

## Procedure

1. Identify actionable vulnerabilities by ecosystem:
   - Use provided advisory context first.
   - If needed, run ecosystem-native security checks.
2. Map each vulnerability to minimum patched versions.
3. Upgrade dependencies with the package manager.
4. Regenerate lock/vendor artifacts with official tooling.
5. Run targeted tests, then broader repository tests.
6. Re-run security checks to confirm remediation.
7. Summarize:
   - advisory/CVE -> dependency -> old/new version
   - validation commands and outcomes
   - risk notes for major upgrades

## Validation Baseline

- `make test`
- Any additional ecosystem checks used for CVE detection/remediation

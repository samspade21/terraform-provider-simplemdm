# SimpleMDM Terraform Acceptance Test Hardening Guide

This guide documents where end-to-end coverage for resources and data sources currently depends on external fixtures and how to close those gaps so GitHub Actions can run as many Terraform apply-style tests as possible with only the `SIMPLEMDM_APIKEY` secret configured.

## What CI runs today

* GitHub Actions executes `go test -v -cover ./provider/` with `TF_ACC=1` and a matrix of Terraform versions. Secrets supply an API key plus numerous optional fixture IDs for data source and resource tests.【F:.github/workflows/test.yml†L74-L117】
* Acceptance tests live under `provider/` and use `testAccPreCheck` to require `TF_ACC` and `SIMPLEMDM_APIKEY`, while `testAccRequireEnv` skips cases that need additional fixtures so day-to-day runs can proceed without them.【F:provider/test_helpers.go†L12-L83】
* The README lists every optional environment variable the suite can honor today, along with a helper script that discovers fixture IDs and prints `gh secret set` commands.【F:README.md†L92-L135】

## Inventory to keep in sync

* Resource and data source registrations in `provider/registry.go` enumerate docs, examples, and test files. Optional acceptance coverage is flagged via `TestsOptional` (e.g., device commands and custom declaration device assignments).【F:provider/registry.go†L8-L160】【F:provider/registry.go†L163-L380】
* `provider/coverage_test.go` enforces that every registered resource/data source has at least one acceptance test file present; new tests must be referenced from the registry to satisfy this check.

## Goals for improved apply-style tests

1. **Rely only on `SIMPLEMDM_APIKEY` whenever feasible.** New tests should create their own dependencies instead of consuming fixture IDs so they can run in forks with a single secret.
2. **Self-provision prerequisites inside Terraform configs.** Prefer creating upstream objects (apps, scripts, profiles, groups) within the test configuration and cleaning them up via `CheckDestroy` rather than reaching for `testAccRequireEnv`.
3. **Minimize real-device assumptions.** Only use environment-dependent skips when hardware or tenant state cannot be mocked (e.g., DDM-capable devices for custom declaration assignments, device-command flows). When a skip is unavoidable, document the exact requirement in the test name and comments.
4. **Add reusable fixtures for complex payloads.** If a resource needs a `.mobileconfig`, custom declaration JSON, or shell script body, check in canonical samples under `provider/testdata/` and reference them from tests so contributors are not blocked by missing local files.
5. **Cover read-only data sources with creation paths.** For data sources that currently expect pre-existing IDs (apps, profiles, script jobs), add companion resources in the same test case to produce those IDs on the fly, then query them with the data source.
6. **Exercise updates, imports, and drift where practical.** Follow existing patterns that apply updates and verify state (e.g., attribute, assignment group, managed config resources) to ensure lifecycle fidelity.

## Implementation checklist for each resource or data source

1. Inspect its entry in `provider/registry.go` to confirm the registered test file path and whether coverage is marked optional. Update the registry when adding new acceptance files so `coverage_test.go` passes.【F:provider/registry.go†L27-L380】
2. Start each acceptance test with `testAccPreCheck(t)` to honor the CI contract, then avoid `testAccRequireEnv` unless a real-world prerequisite cannot be created during the test.【F:provider/test_helpers.go†L12-L34】
3. Build Terraform configs that **create** any prerequisite objects (apps, scripts, managed configs, groups) and reuse their IDs for dependent resources or data sources instead of reading environment variables.
4. For resources requiring binary payloads (custom profiles, mobileconfig uploads, scripts), add deterministic fixture files under `provider/testdata/` and load them via `os.ReadFile` to keep tests reproducible. Include multiple variants when update scenarios are needed.
5. When a genuine external dependency is necessary, wrap that portion in a separate `TestStep` guarded by `testAccRequireEnv` and clearly comment which secret it expects; ensure the base CRUD path still runs with only the API key.
6. Extend tests to cover `ImportState` and `ImportStateVerify` where the API supports it, mirroring patterns in existing resource tests.
7. Keep example directories in sync with new scenarios when they reveal best practices; they are also validated by `coverage_test.go`.

## Areas to prioritize

* **Device-command resource** and **custom declaration device assignment** are currently marked optional; work out API-safe mocks or narrow scenarios that can run without real hardware where possible.【F:provider/registry.go†L69-L115】
* **Data sources tied to existing fleet objects** (device profiles, installed apps, users, script jobs) should receive self-contained tests that create disposable devices or scripts when SimpleMDM API capabilities allow. Otherwise, document the minimal secret needed and keep the test skipped by default.
* **Profile and script payload tests** should ship with sample `.mobileconfig` and script files to unblock contributors. If the API requires specific structure, encode it in fixture files and note any tenant flags required to accept the payloads.

Following this playbook should drive the suite toward comprehensive, reproducible acceptance coverage that GitHub-hosted runners can execute with nothing more than the SimpleMDM API key secret.

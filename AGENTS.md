# AGENTS.md

Repository guidance for epson-exporter.

## Approach

- Stay within the requested scope and preserve unrelated local changes.
- This is a small exporter, not a platform. Prefer direct code and a small dependency surface.
- Simplify and modernize existing code before adding abstractions, compatibility layers, or knobs without a real use case.
- Follow the shared Woodstar tooling baseline while keeping the relaxed Go lint profile appropriate for a small service.

## Repository Map

- Process composition: `cmd/epson_exporter`
- Configuration: `internal/config`
- Device client: `internal/epson`
- Prometheus collection: `internal/exporter`

Keep device transport separate from metric collection. Avoid generic plugin or provider systems.

## Commands

Use Mise tasks as the repository contract.

- Dependencies: `mise run deps`
- Build: `mise run build`
- Tests: `mise run test`
- Lint: `mise run lint`; fixes: `mise run lint-fix`
- Format: `mise run format`; check: `mise run fmt-check`
- Module and workflow checks: `mise run tidy-check`, `mise run workflow-lint`

## Engineering Rules

- Prefer concrete Go types, small consumer-owned interfaces, and wrapped errors.
- Keep metric names and labels stable unless the requested change deliberately changes the public scrape contract.
- Tests use synthetic device responses and local servers; they mustn't call real devices.
- Keep credentials and local configuration out of logs, fixtures, and version control.

## Commits

- Use focused Conventional Commits.
- Don't push, publish images, or contact live devices unless explicitly requested.
- Report checks run, skipped checks, and unresolved failures.

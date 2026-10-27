# AGENT-README

> Orientation for AI coding agents working in this repo. Human docs: README.md; repo rules in `.claude/` (CLAUDE.md, rules/, skills/).

## Type

terraform+go

## Purpose

Terraform Plugin Framework provider letting Terraform manage SigNoz observability resources on both SigNoz Cloud and self-hosted deployments.

## Stack

Go (>= 1.25 to build), terraform-plugin-framework. Generated via [`SigNoz/skaff`](https://github.com/SigNoz/skaff) codegen (OpenAPI spec + `skaff.yml` → `zz_generated_*`). E2E suite in `tests/` (Python + pytest + uv). Gates run through primus.

## Dev mode

```bash
go install .            # build/install the provider binary (README § Building the provider)
make contract-check     # cross-check generated contracts vs a local signoz checkout (SIGOZ_DIR)
```

## Ports / URLs

- none (library/CLI)

## Tests

```bash
make -f "$PRIMUS_HOME/src/make/main.mk" go-test   # primus-pinned go test; run all four go-* gates pre-PR
```

## Gotchas

- Generated, not hand-written: add resources by declaring them in `skaff.yml` and running the skaff pipeline (repo skills `resource-creator` / `spec-syncer`) — never hand-edit `zz_generated_*`.
- skaff and the Go gates run through primus; requires `PRIMUS_HOME` set (use the `primus-setter` skill if missing). Run gates via primus, not bare `go test`/`golangci-lint` — the targets pin CI's versions and flags.
- Requires Terraform >= 1.4 / OpenTofu >= 1.6: nested optional+computed attributes perpetually drift on older CLIs (README § "Why Terraform 1.4").
- Auth via `SIGNOZ_ACCESS_TOKEN` (service-account token); endpoint defaults to `http://localhost:8080`.
- PRs are opened against upstream `SigNoz/terraform-provider-signoz` (`.claude/rules/pull-requests.md`).

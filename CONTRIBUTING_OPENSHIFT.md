# Contributing to the Custom Metrics Autoscaler Operator (OpenShift Downstream)

This document covers contribution guidelines specific to
[openshift/custom-metrics-autoscaler-operator](https://github.com/openshift/custom-metrics-autoscaler-operator),
the OpenShift downstream fork of
[kedacore/keda-olm-operator](https://github.com/kedacore/keda-olm-operator).

This repo is the **operator**. It reconciles a single `KedaController` custom
resource and installs the Custom Metrics Autoscaler **operands** — the KEDA
controller, metrics adapter, and admission webhooks, plus the three optional HTTP
Add-on components — onto a cluster. Changes here frequently have consequences in the
operand repos and vice versa, so read [AGENTS.md](AGENTS.md) for the architecture
before making non-trivial changes.

## Related Resources

| Resource | Link |
|----------|------|
| Upstream repo | [kedacore/keda-olm-operator](https://github.com/kedacore/keda-olm-operator) |
| Operand repo | [openshift/kedacore-keda](https://github.com/openshift/kedacore-keda) |
| HTTP Add-on operand repo | [openshift/kedacore-http-add-on](https://github.com/openshift/kedacore-http-add-on) |
| CI configuration | [openshift/release/.../custom-metrics-autoscaler-operator/](https://github.com/openshift/release/tree/master/ci-operator/config/openshift/custom-metrics-autoscaler-operator) |
| Konflux release pipelines | [openshift/custom-metrics-autoscaler-pipelines](https://github.com/openshift/custom-metrics-autoscaler-pipelines) |
| AI guidance | [AGENTS.md](AGENTS.md) |
| OpenShift docs | [Custom Metrics Autoscaler](https://docs.openshift.com/container-platform/latest/nodes/cma/nodes-cma-autoscaling-custom.html) |
| Release process | [RELEASE-PROCESS.md](RELEASE-PROCESS.md) |
| Approvers | [OWNERS](OWNERS) |

Note the split between the two build systems: **OpenShift CI (Prow)** builds and
tests pull requests, while **Konflux** builds the release artifacts shipped to
customers. A change that passes CI is not automatically a change that ships
correctly; anything touching the Dockerfiles, the CSV, or the bundle layout needs a
thought for the Konflux side too.

## Review and Approval Policy

Every change in every pull request must be understood and approved by two humans.
This can be the PR author and a reviewer, or — if the author used an AI tool and does
not fully understand the contents of the PR — two human reviewers.

**Exception:** PRs authored by deterministic automation tools that are part of our CI
and related systems (whose code has been reviewed by the OpenShift engineering org)
can be merged with a single human review.

Every change should be closely scrutinized for bugs. Our software is complex with
many interdependencies. Review changes from multiple angles:

- **Product architecture**: Does this fit the intended design of the operator, and does it belong in the operator rather than in one of the operands?
- **Security**: Are there new attack surfaces, credential handling issues, or privilege escalations? This operator holds broad RBAC and installs cluster-scoped resources, so RBAC changes deserve particular care.
- **Thread safety**: The reconciler runs alongside background goroutines and separate ConfigMap and Secret controllers that all touch the same `KedaController` status. Are shared resources properly synchronized?
- **Regressions**: Could this break an existing `KedaController` spec field, or change the operand's deployed configuration for users who did not ask for a change?
- **Upgrade safety**: Will an existing installation survive an OLM upgrade to the new CSV? Changes to Deployment selectors, CRDs, or the CSV `replaces` chain have broken upgrades before.
- **Effects on other components**: How does this impact the operand, the metrics API, the HTTP Add-on, or the bundle we ship?

## Upstream Commit Convention

This is a downstream fork. All non-upstream commits must use one of the following
prefixes so that changes are not lost during the next upstream rebase:

- `UPSTREAM: <carry>:` — A change that should be kept (carried) indefinitely, or as long as it makes sense to do so
- `UPSTREAM: <drop>:` — A change that should be discarded during the next rebase cycle
- `UPSTREAM: 1234:` — A change carried until the rebase includes upstream PR #1234

Examples, taken from this repo's actual history:

```text
UPSTREAM: <carry>: Add bundle-previous.Dockerfile.ci for e2e upgrade CI jobs
UPSTREAM: <drop>: Updating and vendoring go modules after an upstream rebase
UPSTREAM: 313: Add e2e upgrade test for KEDA operator
```

Commit prefixes are validated by the `cma-verify-history` CI job, which runs
[`hack/cma-verify-history.sh`](hack/cma-verify-history.sh). A commit is accepted if it
carries one of the prefixes above **or** is already an ancestor of
`kedacore/keda-olm-operator`'s `main` branch. Note that the script's default when run
locally only examines the most recent commit (`HEAD^..HEAD`); CI examines the whole
PR via `PULL_BASE_SHA` and `PULL_PULL_SHA`.

These prefixes are for **commit titles**, not PR titles.

## Upstream-First Policy

New feature work should be directed to the
[upstream project](https://github.com/kedacore/keda-olm-operator). Downstream-only
features are discouraged because of the ongoing cost of maintaining them through
every rebase cycle. If a downstream-only change is necessary, use the
`UPSTREAM: <carry>:` prefix and include a comment in the PR explaining why it cannot
go upstream.

The practical middle ground, and the pattern most of this repo's history follows: open
the PR upstream, then carry it here with the `UPSTREAM: <PR number>:` prefix so it is
dropped automatically once the rebase catches up.

## PR Title Convention

PR titles should be prefixed with a Jira ticket reference:

```text
AUTOSCALE-123: Fix the whatsit in the thingamajig
OCPBUGS-456: Correct nil pointer in KedaController status update
NO-JIRA: Update Go module dependencies
```

The Jira prefix goes in the **PR title**. The upstream commit prefix goes in the
**commit message**.

## PR Workflow

This repo uses [OpenShift CI (Prow)](https://docs.ci.openshift.org/) to gate merges.
The GitHub Actions workflows in [`.github/workflows/`](.github/workflows/) come from
upstream and are not what gates a merge here. PRs merge automatically once all
required tests pass and the correct labels are present.

### Required labels for merge

- `lgtm` — Added by a reviewer via the `/lgtm` command. Any developer from the OpenShift org can add this after reviewing the PR.
- `approved` — Added by an approver listed in the [OWNERS](OWNERS) file via the `/approve` command.
- `verified` — Added by anyone in the OpenShift org, but typically by the PR author, via the `/verified` command. See [Verified Label](#verified-label).

### CI jobs

| Job | What it runs |
|-----|--------------|
| `gofmt` | `make fmt` |
| `govet` | `make vet` |
| `cma-check-all-csv` | `make cma-check-all-csv` — verifies the generated CMA CSVs have not drifted |
| `cma-verify-history` | `hack/cma-verify-history.sh` — validates the `UPSTREAM:` commit prefixes |
| `security` | Snyk scan (optional job) |
| `cma-e2e-aws-ovn` | Installs the bundle via OLM on AWS, applies a `KedaController`, runs the operand e2e suite plus a TLS scan |
| `cma-e2e-gcp-ovn` | The same e2e chain on GCP |
| `cma-e2e-upgrade-aws` | Installs the previous bundle, upgrades to the current one, verifies workloads survive (optional job) |

Most jobs are skipped for documentation-only changes via `skip_if_only_changed`.

### Useful commands

Comment these on the PR:

| Command | Effect |
|---------|--------|
| `/lgtm` | Add the `lgtm` label after reviewing. In repos using [LGTM mode](https://docs.ci.openshift.org/how-tos/creating-a-pipeline/#the-pipeline-required-command), this also triggers E2E and other second-stage tests. |
| `/lgtm cancel` | Remove the `lgtm` label |
| `/approve` | Add the `approved` label (OWNERS approvers only) |
| `/pipeline required` | Manually trigger all required second-stage tests (e.g., E2Es) without waiting for `/lgtm` |
| `/retest` | Re-run all failed tests |
| `/retest-required` | Re-run only the failed required tests |
| `/test <test-name>` | Run a specific test, e.g. `/test cma-e2e-aws-ovn` |
| `/override <context>` | Force a context green — use sparingly, and say why |
| `/hold` | Prevent the PR from being merged |
| `/hold cancel` | Remove the hold and allow merging |
| `/verified` | Mark the PR as verified |
| `/cherry-pick <branch>` | Create a cherry-pick PR to a release branch |

### LGTM mode and E2E tests

Repos enrolled in [LGTM mode](https://docs.ci.openshift.org/how-tos/creating-a-pipeline/#the-pipeline-required-command)
defer second-stage tests (such as E2Es) until the `lgtm` label is applied. This avoids
spending CI resources on PRs that have not been reviewed yet. If you need to run E2Es
before getting `/lgtm` — for example to validate a change before requesting review —
use `/pipeline required`.

### Preventing premature merges

- Add the `WIP:` prefix to the PR title (e.g., `WIP: AUTOSCALE-123: Work in progress`). Prow adds the `do-not-merge/work-in-progress` label automatically.
- Use `/hold` to temporarily block merging while awaiting additional review or testing.

## Test Expectations

PRs should include tests to verify correctness and prevent future regressions:

- **Unit tests**: Required for new logic, bug fixes, and behavior changes. The transform functions in `internal/controller/keda/transform/` and the helpers in `internal/controller/keda/util/` are pure enough to test directly, and are where most new unit tests belong. Run with `make test`.
- **envtest / functionality tests**: Recommended for reconciler behavior. These spin up a real API server through `envtest` and are selected with the custom `-test.type` flag. Run with `make test-functionality`.
- **E2E tests**: Expected for new features and significant behavior changes. `make e2e-test` runs the smoke suite in `test/e2e/` against a cluster that already has the operator installed via OLM.
- **Upgrade tests**: Required whenever you touch the CSV, the bundle, CRDs, or Deployment selectors. See the sequence below.

Note that the test targets depend on `generate fmt vet`, so running them may modify
files in your working tree. Check `git status` afterwards.

### OLM upgrade test sequence

The upgrade test is five targets run in order. Three of them do the OLM plumbing
(`e2e-olm-upgrade-*`) and two are the Go tests that run inside it
(`e2e-upgrade-test-*`, which are `TestKedaUpgradeSetup` and
`TestKedaUpgradeVerify`). They interleave, so the order matters:

```bash
make e2e-olm-upgrade-build     # build the operator image plus the previous and current bundles
make e2e-olm-upgrade-install   # install the previous version via OLM
make e2e-upgrade-test-pre      # deploy workloads under the previous version
make e2e-olm-upgrade-apply     # operator-sdk run bundle-upgrade to the current bundle
make e2e-upgrade-test-post     # verify the workloads survived the upgrade
```

Run `make e2e-olm-cleanup` afterwards to tear the OLM install down.

The "previous" version is auto-detected from the latest upstream GitHub release, and
the current bundle's CSV is relabelled with a synthetic `1000.0.0` version so OLM sees
a newer entry to upgrade to. `.github/workflows/pr-tests.yml` runs this exact order
using the `-ci` variants of the two test targets, which add `gotestsum` output.

## Verified Label

Use `/verified` to indicate that changes have been verified. Examples:

```text
/verified
/verified by cma-e2e-aws-ovn
/verified by unit tests
/verified by E2Es
/verified later
```

## Generated Code

The following are generated and should never be hand-edited:

| File(s) | Generator | Regenerate with |
|---------|-----------|-----------------|
| `api/keda/v1alpha1/zz_generated.deepcopy.go` | controller-gen | `make generate` |
| `config/crd/bases/keda.sh_kedacontrollers.yaml` | controller-gen, from kubebuilder markers | `make manifests` |
| `config/rbac/role.yaml` | controller-gen, from `+kubebuilder:rbac` markers | `make manifests` |
| `bundle/manifests/`, `bundle/metadata/` | operator-sdk + kustomize | `make bundle` |
| `keda/<version>/manifests/cma.v<version>.clusterserviceversion.yaml` | `hack/cma-generate-csv.sh` | `hack/cma-generate-csv.sh <version>` |
| `resources/keda.yaml` | Imported verbatim from an upstream KEDA release | `hack/relprep.sh <version>` |
| `resources/keda-http-addon.yaml` | Imported verbatim from an upstream HTTP Add-on release | `hack/http-add-on-relprep.sh <version>` |
| `vendor/` | Go tooling | `go mod tidy && go mod vendor` |

The other operand CRDs under `config/crd/bases/` (`ScaledObject`, `ScaledJob`,
`TriggerAuthentication`, `CloudEventSource`, `HTTPScaledObject`, and so on) are copied
in from upstream operand releases by the relprep scripts, not generated from Go types
in this repo. Only `keda.sh_kedacontrollers.yaml` comes from controller-gen.

Always commit generated file changes in the same PR as the source changes, and put
`vendor/` churn in a separate commit from logic changes.

## Development Quick Reference

| Task | Command |
|------|---------|
| Build the manager binary | `make build` |
| Run the operator locally against your kubeconfig | `make install && make run` |
| Run unit tests | `make test` |
| Run functionality tests (envtest) | `make test-functionality` |
| Run OLM deployment tests | `make test-deployment` |
| Run audit-flag tests | `make test-audit` |
| Run linters | `make lint` |
| Format code | `make fmt` |
| Run go vet | `make vet` |
| Generate deepcopy code | `make generate` |
| Generate CRD and RBAC manifests | `make manifests` |
| Generate the OLM bundle | `make bundle` |
| Verify the CMA CSVs are in sync | `make cma-check-all-csv` |
| Install the CRDs only | `make install` |
| Deploy the operator with kustomize | `make deploy` |
| Deploy the operator via OLM | `make deploy-olm` |
| Set up an OLM install for e2e | `make e2e-olm-setup` |
| Run e2e smoke tests | `make e2e-test` |
| Run the OLM upgrade e2e sequence | Five ordered targets — see [OLM upgrade test sequence](#olm-upgrade-test-sequence) |
| Tear down the OLM install | `make e2e-olm-cleanup` |
| List all targets | `make help` |

The linter target is `make lint`, and it downloads `golangci-lint` when needed. There is also no
`make verify` target; check for generated-file drift with `git diff --exit-code` as shown below.

`golangci-lint` and `pre-commit` are the two tools worth installing locally. The
`pre-commit` hooks (`pre-commit install`) run `go fmt`, `golangci-lint`, a
whitespace/EOF fixer, a private key detector, and `doctoc` for the README.

## Pre-Submit Checklist

Before requesting review:

1. `make build` — verify the code compiles
2. `make test` — run unit tests
3. `make lint` — run linters
4. `make manifests generate && git diff --exit-code` — ensure generated files are up to date
5. `make cma-check-all-csv` — ensure the CMA CSVs have not drifted (required if you touched the CSV or `hack/cma-generate-csv.sh`)
6. Confirm every commit has an `UPSTREAM:` prefix and the PR title has a Jira prefix
7. Review your diff for secrets, credentials, debug code, and unintended `config/` churn from `make bundle` or `make deploy`
8. Address any [CodeRabbit](https://coderabbit.ai/) review feedback, as a courtesy to the human reviewer who follows. You do not need to accept every suggestion — replying with an explanation of why you are not acting on one is perfectly fine. The goal is to clear the straightforward issues so human reviewers can focus on the substance of the change.

## Code Style

- Run `make fmt` before committing, and `make lint` to catch the rest. The configuration lives in [`.golangci.yml`](.golangci.yml).
- Import ordering is enforced by `gci`: standard library, then external packages, then `github.com/kedacore/keda-olm-operator` (separated by blank lines).
- Follow Go conventions for error strings: lowercase, no trailing punctuation, wrapped with `fmt.Errorf("context: %w", err)`.
- Use structured logging via `logr`: constant messages with key-value pairs in lowerCamelCase, rather than formatted strings.
- Use `deny_list` and `allow_list` rather than the older colour-based terms for the same concepts. The `language-matters` pre-commit hook rejects the latter, including in comments and documentation.
- Keep the Go module path as `github.com/kedacore/keda-olm-operator`. It is the upstream module path and the `canonical_go_repository` in our CI configuration; renaming it would break both.

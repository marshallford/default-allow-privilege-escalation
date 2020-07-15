# Kubernetes Mutating Webhook for Defaulting AllowPrivilegeEscalation

[![Build Status](https://github.com/marshallford/default-allow-privilege-escalation/workflows/CI/badge.svg)](https://github.com/marshallford/default-allow-privilege-escalation/actions?query=workflow%3ACI)
[![Go Report](https://goreportcard.com/badge/github.com/marshallford/default-allow-privilege-escalation)](https://goreportcard.com/report/github.com/marshallford/default-allow-privilege-escalation)
[![Codecov](https://codecov.io/gh/marshallford/default-allow-privilege-escalation/branch/master/graphs/badge.svg)](https://codecov.io/github/marshallford/default-allow-privilege-escalation)
[![License](https://img.shields.io/github/license/marshallford/default-allow-privilege-escalation)](/LICENSE)

Controls the nil behavior of the field `allowPrivilegeEscalation` in the [`SecurityContext`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#securitycontext-v1-core) object. Useful in cases where the PSP admission controller isn't enabled or available. With PSP, this behavior is managed the `*bool` type field named [`defaultAllowPrivilegeEscalation`](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/auth/no-new-privs.md#pod-security-policy-changes) in a Pod Security Policy resource.

**TODO:**

- [x] find a better way to test Fiber handlers
- [x] tests for config and health packages
- [ ] webhook should self-manage CA bundle
- [x] Github Actions with test and coverage badges
- [x] improve makefile
- [ ] release CI with tagging
- [ ] publish container image
- [ ] flesh out deploy yaml, add Kustomize support
- [ ] provide install instructions
- [ ] docs showing behavior

## 🏁 Quickstart

TODO

## ⚙️ Configure

TODO

## 🤖 Hack

### Test

```shell
make test
make coverage
```

### Build

```shell
make build
make docker-build # builds container image
```

### Run

```shell
make run
make docker-run # runs container image
```

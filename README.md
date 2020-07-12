# Kubernetes Mutating Webhook for defaulting AllowPrivilegeEscalation

Controls the nil behavior of the field `allowPrivilegeEscalation` in the [`SecurityContext`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#securitycontext-v1-core) object. Useful in cases where the PSP admission controller isn't enabled or available. With PSP, this behavior is managed the `*bool` type field named [`defaultAllowPrivilegeEscalation`](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/auth/no-new-privs.md#pod-security-policy-changes) in a Pod Security Policy resource.

**TODO:**

- [ ] find a better way to test Fiber handlers
- [ ] tests for config and health packages
- [ ] webhook should self-manage CA bundle
- [ ] Github Actions with test and coverage badges
- [ ] improve makefile
- [ ] release CI with tagging
- [ ] publish container image
- [ ] flesh out deploy yaml, add Kustomize support
- [ ] provide install instructions
- [ ] docs showing behavior

## Contribute

### Test

```shell
make test
```

### Build

```shell
make build
make docker-build # container image
```

### Run

```shell
make run
```

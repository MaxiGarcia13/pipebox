# Vue reactivity — no unnecessary memoize

Do not wrap cheap, synchronous derives in `computed` or `watch`. Prefer plain `const`, template expressions, or a real derived `computed` only when it earns its keep.

## Prefer

| Use                    | When                                                                       |
| ---------------------- | -------------------------------------------------------------------------- |
| Plain `const` / inline | Cheap sync value from stable props (paths, static variant, maps)           |
| Template expression    | Cheap derive from props that can change (`formatX(prop)`, class ternaries) |
| `computed`             | Reactive deps + reuse, or non-trivial work worth caching                   |
| `watch`                | Side effects only (fetch, DOM, navigate, reset unrelated state)            |

## Avoid

| Avoid                                    | Why                                                                       |
| ---------------------------------------- | ------------------------------------------------------------------------- |
| `computed(() => buildPath(props.id))`    | No caching benefit; use `const href = buildPath(props.id)`                |
| `watch(x, v => { y.value = f(v) })`      | That is derived state — use `computed` (or a plain const if not reactive) |
| `watch` to mirror props into local `ref` | Prefer using the prop (or `toRef`) directly                               |

```typescript
// ❌ BAD — memoize a cheap string
const logHref = computed(() => buildStepLogPath(props.step.id));

// ✅ GOOD
const logHref = buildStepLogPath(props.step.id);

// ❌ BAD — watch as derived state
watch(workflowQuery, (value) => {
  isEmpty.value = !value.trim();
});

// ✅ GOOD
const isEmpty = computed(() => !workflowQuery.value.trim());

// ✅ GOOD — watch for a side effect
watch(repoPath, (value) => {
  void loadWorkflows(value);
});
```

`computed` is fine for UI gates and cascades (`showSkeleton`, `errorMessage`, filter option lists). Do not add it by habit.

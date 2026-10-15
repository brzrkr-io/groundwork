# Labels and selectors: selecting on the server

A big cluster holds thousands of Pods. Instead of downloading them all and
filtering in Go, you hand the selector to the API server and get back only
the Pods you asked for.

**New here:** `List` with `metav1.ListOptions{LabelSelector: ...}`, `metav1.NamespaceAll`, a Service's selector vs its own labels.

## Learn

### Recap

A selector such as `app=checkout,tier=web` picks objects whose labels match.
`labels.SelectorFromSet(set).String()` turns a map into selector text. As in
*The clientset: get, create and delete*, your functions take a
`kubernetes.Interface`, and `fake.NewClientset(objects...)` plays the API
server in tests.

### Selecting on the server

`List` takes a `metav1.ListOptions`. Put selector text in its `LabelSelector`
field and the API server does the filtering, like `kubectl get -l`:

```go
secrets, err := cs.CoreV1().Secrets("billing").List(ctx, metav1.ListOptions{
	LabelSelector: "owner=vault",
})
if err != nil {
	return err
}
for _, s := range secrets.Items { // secrets is a *corev1.SecretList
	fmt.Println(s.Name)
}
```

> **Why is it written like this?** `.Items` is required: `List` returns a
> list *object* (with metadata of its own), and the objects are in its `Items`
> slice. Naming the field (`LabelSelector:`) in the struct literal is
> convention. Fields you don't name stay empty.

The API server returns lists sorted by namespace, then name. The fake does
the same and honours `LabelSelector` too.

An empty `LabelSelector` means "no filter": you get everything. To search
every namespace, like `kubectl get pods -A`, pass `metav1.NamespaceAll` as the
namespace. It is just `""`, but the name says what you mean.

### A Service's two sets of labels

| Field | What it describes |
|---|---|
| `svc.Labels` (`metadata.labels`) | the Service itself |
| `svc.Spec.Selector` (`spec.selector`) | the Pods it sends traffic to |

They often look alike, so they are easy to mix up. A Service with no selector
sends traffic to **no** Pods; someone must wire up its endpoints by hand.

`len` works on a `nil` map and returns `0`.

## Your task

You're on call for the `shop` cluster and want two quick answers: "which Pods
are behind this Service?" and "how many Pods does this team run in each
environment?".

### 1. Fix `PodsForService`

**Goal:** make `PodsForService` list the Pods picked by the Service's selector, and return `nil, nil` for a Service with no selector.

A teammate reports two bugs:

- It lists Pods using the Service's own labels (`svc.Labels`). It must use
  `svc.Spec.Selector`.
- For a Service with no selector it returns every Pod in the namespace. It
  must return `nil, nil` without listing anything.

Keep the rest: names in the order `List` returns them, and the
`get service ...` error (wrapped with `%w`) when the Service doesn't exist.

```go
// Service shop/checkout: labels   app=checkout,team=payments
//                        selector app=checkout,tier=web
PodsForService(ctx, cs, "shop", "checkout")
// => []string{"checkout-web-a", "checkout-web-b"}, nil
PodsForService(ctx, cs, "shop", "legacy-db") // no selector
// => nil, nil
```

**Think it through:**
- Both bugs are on or just above the `selector :=` line.
- What does an empty selector match, and so where must you return early?

### 2. Finish `CountByEnv`

**Goal:** make `CountByEnv` list the Pods labelled `team=<team>` in **all** namespaces and count them by their `env` label.

The counting loop already works and skips Pods without `env`. It is missing
the part that fetches the Pods:

- Build the selector text from the `team` parameter, with the fixed key `team`.
- List with `LabelSelector` and `metav1.NamespaceAll`. If `List` fails, return
  `nil` and the error.
- Loop over the listed Pods (`.Items`) instead of the empty `pods` slice. Then
  delete that slice and the unused `corev1` import.
- When the team has no Pods, return an empty map (not `nil`).

```go
// team=payments: 4 Pods with env=prod, 1 with env=staging,
// and 1 with no env label, spread over three namespaces.
CountByEnv(ctx, cs, "payments")
// => map[string]int{"prod": 4, "staging": 1}, nil
```

**Think it through:**
- `PodsForService` already builds selector text from a map. How would you build a one-entry map from `team`?
- What namespace argument means "every namespace"?

## Learn more

- [Kubernetes: Labels and Selectors, API](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api)
- [Kubernetes: Services without selectors](https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors)
- [pkg.go.dev: metav1.ListOptions](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ListOptions)

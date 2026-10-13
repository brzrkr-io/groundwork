# Labels and selectors: selecting on the server

A big cluster holds thousands of Pods. Instead of downloading them all and
filtering in Go, you hand the selector to the API server and get back only
the Pods you asked for.

## Learn

### Recap

Labels are key/value tags such as `app: checkout`, and a selector such as
`app=checkout,tier=web` picks the objects whose labels match.
`labels.SelectorFromSet(set).String()` turns a map into selector text. As in
*The clientset: get, create and delete*, your functions take a `kubernetes.Interface` and reach each
kind of object through a typed client, such as
`cs.CoreV1().Services(ns).Get(ctx, name, metav1.GetOptions{})`. In tests,
`fake.NewClientset(objects...)` plays the API server.

### Selecting on the server

`List` takes a `metav1.ListOptions`. Put selector text in its `LabelSelector`
field and the API server does the filtering, exactly like
`kubectl get pods -l app=checkout`:

```go
pods, err := cs.CoreV1().Pods("shop").List(ctx, metav1.ListOptions{
	LabelSelector: "app=checkout,tier=web",
})
if err != nil {
	return err
}
for _, pod := range pods.Items { // pods is a *corev1.PodList
	fmt.Println(pod.Name)
}
```

The API server returns lists sorted by namespace, then by name. The fake
clientset does the same, and it honours `LabelSelector` too.

An empty `LabelSelector` means "no filter": you get every Pod. To search
every namespace at once, like `kubectl get pods -A`, pass
`metav1.NamespaceAll` as the namespace. It is just the empty string `""`, but
the name says what you mean.

### A Service's two sets of labels

A Service has labels of its own in `metadata.labels`, which describe the
Service itself. The labels it sends traffic to are in `spec.selector`. These
often look alike, which makes them easy to mix up. A Service with no
selector at all sends traffic to no Pods (someone wires it up by hand).

## Your task

You're on call for the `shop` cluster and want two quick answers: "which
Pods are behind this Service?" and "how many Pods does this team run in each
environment?".

### 1. Fix `PodsForService`

`PodsForService(ctx, cs, namespace, name)` gets the Service and lists the
Pods behind it, but a teammate reports two bugs:

- It lists Pods using the Service's own labels (`svc.Labels`). It must use the
  Service's selector, `svc.Spec.Selector`.
- For a Service with no selector it returns every Pod in the namespace. It
  must return `nil, nil` without listing anything.

Keep the rest as it is: names in the order the List call returns them, and
the "get service" error (wrapped with `%w`) when the Service doesn't exist.

```go
// Service shop/checkout: labels app=checkout,team=payments
//                        selector app=checkout,tier=web
PodsForService(ctx, cs, "shop", "checkout")
// => []string{"checkout-web-a", "checkout-web-b"}, nil
PodsForService(ctx, cs, "shop", "legacy-db") // no selector
// => nil, nil
```

### 2. Finish `CountByEnv`

`CountByEnv(ctx, cs, team)` already counts Pods by their `env` label and
skips Pods that have none. It is missing the part that fetches the Pods.
List the Pods labelled `team=<team>` in **all** namespaces, using
`LabelSelector` and `metav1.NamespaceAll`. If the List call fails, return
`nil` and the error. Then loop over the listed Pods (`.Items`) instead of the
empty `pods` slice. When the team has no Pods, return an empty map.

```go
// team=payments: 4 Pods with env=prod, 1 with env=staging,
// and 1 with no env label, spread over three namespaces.
CountByEnv(ctx, cs, "payments")
// => map[string]int{"prod": 4, "staging": 1}, nil
```

## Learn more

- [Kubernetes: Labels and Selectors, API](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api)
- [Kubernetes: Services without selectors](https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors)
- [pkg.go.dev: metav1.ListOptions](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ListOptions)

package reconcilier

func (r *Reconciler) Reconcile(key string) error {
    desired := r.store.GetDesired(key)
    actual := r.store.GetActual(key)

    ops := diff(desired, actual)

    for _, op := range ops {
        r.actuator.Execute(op)
    }

    r.client.ReportState(key, actual)
}
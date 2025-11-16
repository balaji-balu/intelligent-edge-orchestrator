// -------------------------
// Unit-test skeleton (user to expand)
// -------------------------

func TestHashSpec(t *testing.T) {
    spec := ComponentSpec{Name: "api", Image: "api:v1"}
    h, err := hashComponentSpec(spec)
    if err != nil { t.Fatal(err) }
    if h == "" { t.Fatalf("expected hash") }
}

func TestDiffInstall(t *testing.T) {
    desired := DeploymentSpec{DeploymentID: "d1", Components: []ComponentSpec{{Name: "api", Image: "api:v1"}}}
    actual := DeploymentState{DeploymentID: "d1", Components: []ComponentState{}}
    ops, err := diff(desired, actual)
    if err != nil { t.Fatal(err) }
    if len(ops) != 1 || ops[0].Type != OpInstall { t.Fatalf("expected install op") }
}

func TestReconcileFlow(t *testing.T) {
    // implement fake store + fake actuator + fake reporter and assert calls
}

// -------------------------
// Example Actuator stub (for tests / local)
// -------------------------

type FakeActuator struct{}

func (f *FakeActuator) Execute(op Operation) error {
	// Simulate action latency
	time.Sleep(50 * time.Millisecond)
	switch op.Type {
	case OpInstall:
		fmt.Printf("[FAKE] install %s on %s\n", op.Component.Name, op.TargetNode)
	case OpUpdate:
		fmt.Printf("[FAKE] update %s on %s\n", op.Component.Name, op.TargetNode)
	case OpRemove:
		fmt.Printf("[FAKE] remove %s on %s\n", op.Component.Name, op.TargetNode)
	case OpNoOp:
		// nothing
	default:
		return errors.New("unknown op")
	}
	return nil
}

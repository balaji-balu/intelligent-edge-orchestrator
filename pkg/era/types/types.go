package types

type Spec struct {
    Name      string
    Version   string
    Image     string
    WasmFile  string
    Args      []string
}

cat internal/era/plugins/plugin.go
cat internal/era/lifecycle/controller.go
cat internal/era/reporter/status_reporter.go
cat internal/era/runtimemanager/manager.go
cat pkg/era/types/types.go
cat pkg/era/*.go  
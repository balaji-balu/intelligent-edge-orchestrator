package main

import (
    "fmt"
    "github.com/balaji-balu/margo-hello-world/internal/era/plugins"
    "github.com/balaji-balu/margo-hello-world/internal/era/runtimemanager"
    "github.com/balaji-balu/margo-hello-world/pkg/era"
)

func main() {
    plugin := plugins.NewRuntimePlugin()

    rm := runtimemanager.NewRuntimeManager(plugin)

    comp := era.ComponentSpec{
        Name:     "hello",
        Runtime:  "wasm",
        Artifact: "hello.wasm",
    }

    fmt.Println("Deploying component...")
    rm.Deploy(comp)

    fmt.Println("Status:", rm.GetStatus("hello"))
}

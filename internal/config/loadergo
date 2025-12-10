package config

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/providers/env"
)

var k = koanf.New(".")

type Loader struct {
    AppName string
    Env     string
}

func New() *Loader {
    l := &Loader{
        AppName: os.Getenv("APP_NAME"),
        Env:     os.Getenv("APP_ENV"),
    }

    if l.AppName == "" {
        panic("APP_NAME must be set (co, lo, era, edgectl)")
    }
    if l.Env == "" {
        l.Env = "development"
    }

    return l
}

func (l *Loader) Load(out interface{}) error {
    configPath := filepath.Join("configs", l.AppName, l.Env+".yaml")

    err := k.Load(file.Provider(configPath), yaml.Parser())
    if err != nil {
        return fmt.Errorf("error loading config %s: %w", configPath, err)
    }

    // Load ENV overrides: APP_<SECTION>_<KEY>
    // Example: APP_NATS_PASSWORD, APP_LOG_LEVEL, APP_CONTAINERD_SOCKET
    k.Load(env.Provider("APP_", ".", func(s string) string {
        return s
    }), nil)

    return k.Unmarshal("", out)
}

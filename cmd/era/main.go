package main

import (
    //"fmt"
    "io"
    "encoding/json"
    "os"
    "strings"
    "net/http"
    "path/filepath"
    "bytes"
    "errors"
    "runtime"
    "github.com/google/uuid"
    "go.uber.org/zap"

    "github.com/balaji-balu/margo-hello-world/internal/era/runtimemgr"
    "github.com/balaji-balu/margo-hello-world/pkg/logx"
    "github.com/balaji-balu/margo-hello-world/internal/config"
    "github.com/balaji-balu/margo-hello-world/internal/natsbroker"
    "github.com/balaji-balu/margo-hello-world/internal/era/heartbeat"
    _ "github.com/balaji-balu/margo-hello-world/internal/era/plugins/containerd"
)
// sudo ctr -n era containers ls
// sudo ctr -n era tasks ls

// sudo ctr -n era tasks kill --signal KILL edge-ai-sample
// sudo ctr -n era tasks delete edge-ai-sample
// sudo ctr -n era containers delete edge-ai-sample
// func main() {
//     plugin := plugins.NewRuntimePlugin()

//     rm := runtimemanager.NewRuntimeManager(plugin)

//     comp := era.ComponentSpec{
//         Name:     "hello",
//         Runtime:  "wasm",
//         Artifact: "hello.wasm",
//     }

//     fmt.Println("Deploying component...")
//     rm.Deploy(comp)

//     fmt.Println("Status:", rm.GetStatus("hello"))
// }

type EraConfig struct {
    Log struct {
        Level  string `koanf:"level"`
        Format string `koanf:"format"`
    } `koanf:"log"`

    NATS struct {
        URL      string `koanf:"url"`
        Username string `koanf:"username"`
        Password string `koanf:"password"`
    } `koanf:"nats"`

    LO struct {
        URL     string `koanf:"url"`
    }
}

var log *zap.SugaredLogger
func Init() {

}

func main() {

    ls, err := InitERAStorage() //loadOrCreateHostID(getBaseDir("ERA"), "host_id")

    logx.Init(logx.Options{
        Env:     os.Getenv("APP_ENV"),     // dev / prod
        Version: "0.1.0",
    })    
    log = logx.New("era")
    log.Infow("ERA starting", "pid", os.Getpid())

    log.Infow("ls", "", ls)
    
    loader := config.New()
    var cfg EraConfig
    if err := loader.Load(&cfg); err != nil {
        log.Errorw("config load err", err)
    }    
    log.Infow("Loaded ERA config:", "config", cfg)    

    log.Infow("📡 Connecting to ","NATS at", cfg.NATS.URL)
	nb, err := natsbroker.New(cfg.NATS.URL)
	if err != nil {
		log.Errorf("❌ Failed to connect to NATS.","err:", err)
        return
	}
    siteID, err := register(cfg.LO.URL, ls.HostID)
    if err != nil {
        log.Errorf("❌ Unable to Register with LO","err:", err)
        return
    }
    log.Infow("LO", "siteid", siteID)
    
    heartbeat.StartHeartbeat(nb, log, siteID, ls.HostID)
    // Pass log into your DI / top-level orchestrator

    // comp := edgeruntime.ComponentSpec{
    //     Name:     "edge-ai-sample",
    //     Runtime:  "containerd",
    //     Artifact: "ghcr.io/edge-orchestration-platform/edge-ai-sample:74fb8f5c0bcdeecb53685605a1c30889b33601b6",
    // }
    era := runtimemgr.NewRuntimeManager("containerd", nb, log)
    era.LoActionDispatcher(siteID, ls.HostID)

    // log.Infow("Deploy status", "", era.Deploy(comp))

    // // get the status
    // log.Infow("container status", "", era.GetStatus(comp.Name))

    // time.Sleep(1 * time.Minute)

    // // stop the container
    // status := era.Stop(comp.Name)
    // log.Infow("stop", "status", status)
    // status = era.Delete(comp.Name )
    // log.Infow("Delete", "status", status)
    // log.Infow("container status", "", era.GetStatus(comp.Name))    
    select{}
}

func register(loURL, hostID string) (string, error) {
    // Prepare payload
    payload := map[string]string{
        "host_id": hostID,
    }

    b, err := json.Marshal(payload)
    if err != nil {
        log.Errorw("failed to serialize payload:", "err", err)
        return "", err
    }

    // Create request
    req, err := http.NewRequest("POST", loURL+"/register", bytes.NewBuffer(b))
    if err != nil {
        log.Errorw("failed to create request:","err", err)
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")

    // Make request
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Errorw("register request failed: ","err", err)
        return "", err
    }
    defer resp.Body.Close()

    // Check for non-OK status
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        log.Errorw("LO returned", "statuscode", resp.StatusCode, "body", string(body))
        return "", errors.New("returned non ok status") 
    }

    // Parse LO's response (siteID)
    var siteID string
    if err := json.NewDecoder(resp.Body).Decode(&siteID); err != nil {
        log.Errorw("failed to decode LO response: ","err", err)
        return "", err
    }

    return siteID, nil
}


func loadOrCreateID(baseDir, name string) (string, error) {
    idPath := filepath.Join(baseDir, name)

    if data, err := os.ReadFile(idPath); err == nil {
        id := strings.TrimSpace(string(data))
        if id != "" {
            return id, nil
        }
    }

    id := uuid.New().String()

    os.MkdirAll(baseDir, 0755)
    os.WriteFile(idPath, []byte(id), 0644)

    return id, nil
}


type ERAStorage struct {
    BaseDir string
    HostID  string
}

func InitERAStorage() (*ERAStorage, error) {
    baseDir := ERABaseDir() // cross-platform version same as LOBaseDir

    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return nil, err
    }

    hostID, err := loadOrCreateID(baseDir, "host_id")
    if err != nil {
        return nil, err
    }

    return &ERAStorage{
        BaseDir: baseDir,
        HostID:  hostID,
    }, nil
}

func ERABaseDir() string {
    if os.Getenv("APP_ENV") == "development" {
        home, _ := os.UserHomeDir()
        return filepath.Join(home, ".era")
    }

    switch runtime.GOOS {
    case "windows":
        return filepath.Join(os.Getenv("ProgramData"), "ERA")

    case "darwin":
        return filepath.Join("/Library/Application Support", "ERA")

    default:
        return "/var/lib/era"
    }
}

package e2e

import (
    "os"
	"strings"
    "os/exec"
    "path/filepath"
    "testing"
)

var edgectlBin string

func TestMain(m *testing.M) {
    // Expand ~
    home, _ := os.UserHomeDir()
    edgectlPath := filepath.Join(
        home,
        "edge-orchestration-workspace/prototypes/edgectl",
    )

    // Output binary path
    edgectlBin = filepath.Join(os.TempDir(), "edgectl-test-bin")
    os.MkdirAll(edgectlBin, 0755)
    edgectlBin = filepath.Join(edgectlBin, "edgectl")

    // Build the CLI
    cmd := exec.Command(
        "go", "build",
        "-o", edgectlBin,
        ".",
    )
    cmd.Dir = edgectlPath

    out, err := cmd.CombinedOutput()
    if err != nil {
        panic("failed to build edgectl:\n" + string(out))
    }

    // Tell CLI where to read .env from
    os.Setenv("EDGECTL_ENV_PATH",
        "/home/balaji/edge-orchestration-workspace/prototypes/edgectl/.env",
    )

    // Run tests
    os.Exit(m.Run())
}

func runCLI(t *testing.T, args ...string) (string, error) {
    t.Helper()
    cmd := exec.Command(edgectlBin, args...)
    out, err := cmd.CombinedOutput()
    return string(out), err
}


func initGit(t *testing.T) {
    exec.Command("git", "init").Run()
    exec.Command("git", "config", "user.email", "test@test.com").Run()
    exec.Command("git", "config", "user.name", "tester").Run()

    // Add an initial commit (required for many Git ops)
    os.WriteFile("README.md", []byte("test repo"), 0644)
    exec.Command("git", "add", ".").Run()
    exec.Command("git", "commit", "-m", "init").Run()
}

func TestDeploy_E2E(t *testing.T) {
    // Create isolated workspace
    dir := t.TempDir()
    old := os.Getenv("PWD")
    os.Chdir(dir)
    defer os.Chdir(old)

    // Init git repository
    initGit(t)

    // 1. Create or add an app first
    // (co deploy expects app exists)
    out, err := runCLI(t,
        "co", "add", "app",
        "--name", "retail/pos/1.1",
        "--artifact", "file://testdata/pos",
    )
    if err != nil {
        t.Fatalf("add app failed: %v\nout: %s", err, out)
    }

    // 2. Deploy the app
    out, err = runCLI(t,
        "co", "deploy",
        "--app", "retail/pos/1.1",
        "--site", "store-121",
    )
    if err != nil {
        t.Fatalf("deploy failed: %v\nout: %s", err, out)
    }

    // 3. Validate output
    if !strings.Contains(out, "Deployment created") &&
       !strings.Contains(out, "Success") {
        t.Fatalf("unexpected deploy output: %s", out)
    }

    // 4. Verify desired state file got generated
    dsFile := "desiredstate/store-121/retail-pos-1.1.yaml"
    if _, err := os.Stat(dsFile); err != nil {
        t.Fatalf("desired state not generated: expected %s", dsFile)
    }

    // 5. Verify commit happened in Git
    gitLog, _ := exec.Command("git", "log", "--oneline", "-1").CombinedOutput()
    if !strings.Contains(string(gitLog), "deploy") &&
       !strings.Contains(string(gitLog), "desiredstate") {
        t.Fatalf("deploy did not create a git commit: %s", gitLog)
    }
}


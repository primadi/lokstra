package lokstra

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/primadi/lokstra/core/annotation"
	"github.com/primadi/lokstra/core/deploy"
)

type RunMode string

const (
	RunModeProd  RunMode = "prod"
	RunModeDev   RunMode = "dev"
	RunModeDebug RunMode = "debug"
)

var (
	Mode        RunMode
	childEnvKey = "LOKSTRA_CHILD"
)

// Bootstrap initializes Lokstra environment and regenerates routes if needed.
// It must be called at the very beginning of main().
func Bootstrap() {
	// 1️⃣ Check for --generate-only flag
	generateOnly := false
	for _, arg := range os.Args {
		if arg == "--generate-only" {
			generateOnly = true
			break
		}
	}

	if generateOnly {
		fmt.Println("[Lokstra] Running in GENERATE-ONLY mode")
		fmt.Println("[Lokstra] Force rebuilding all generated code...")

		// Force rebuild by deleting all cache files first
		if err := deleteAllCacheFiles(); err != nil {
			fmt.Println("[Lokstra] Warning: failed to delete cache files:", err)
		}

		// Run autogen
		_, err := runAutoGen()
		if err != nil {
			fmt.Println("[Lokstra] Autogen failed:", err)
			os.Exit(1)
		}

		fmt.Println("[Lokstra] ✅ Code generation completed successfully")
		os.Exit(0)
	}

	// 2️⃣ Detect mode
	Mode = detectRunMode()
	fmt.Printf("[Lokstra] Environment detected: %s\n", strings.ToUpper(string(Mode)))

	// 3️⃣ Prevent infinite loop
	if os.Getenv(childEnvKey) == "1" {
		// fmt.Println("[Lokstra] Child process detected — skipping bootstrap autogen.")
		return
	}

	// 4️⃣ If prod, just continue
	if Mode == RunModeProd {
		// fmt.Println("[Lokstra] Production mode — skipping autogen.")
		return
	}

	// 5️⃣ Run autogen
	codeChanged, err := runAutoGen()
	if err != nil {
		fmt.Println("[Lokstra] Autogen failed:", err)
		os.Exit(1)
	}

	// 6️⃣ Relaunch using correct method
	switch Mode {
	case RunModeDebug:
		if codeChanged {
			relaunchWithDlv()
		}
	case RunModeDev:
		relaunchWithGoRun()
	default:
		fmt.Println("[Lokstra] Unknown mode — continuing normally.")
	}
}

// detectRunMode inspects runtime and env to detect how the app was started.
func detectRunMode() RunMode {
	exe, err := os.Executable()
	if err != nil {
		fmt.Println("[Lokstra] Warning: cannot get executable path:", err)
		return RunModeProd
	}

	exePath := filepath.ToSlash(exe)
	exeName := filepath.Base(exe)
	deploy.LogDebug("[Lokstra] Executable: %s\n", exePath)

	// 1️⃣ Check if running under Delve debugger
	// Delve wraps the binary with __debug_bin
	if strings.Contains(exeName, "__debug_bin") {
		fmt.Println("[Lokstra] Detected: Delve debugger (debug binary)")
		return RunModeDebug
	}

	// 2️⃣ Check environment variables for debugger presence
	// VSCode sets these when debugging, Delve also sets DLV_* vars
	for _, e := range os.Environ() {
		key := strings.Split(e, "=")[0]

		// Delve environment variables
		if strings.HasPrefix(key, "DLV_") {
			fmt.Println("[Lokstra] Detected: Delve environment variable:", key)
			return RunModeDebug
		}

		// VSCode debugger detection (more specific check)
		if key == "VSCODE_DEBUGGER_RUNTIME_TYPE" ||
			key == "VSCODE_DEBUG_PROTOCOL_VERSION" {
			fmt.Println("[Lokstra] Detected: VSCode debugger environment")
			return RunModeDebug
		}
	}

	// 3️⃣ Check if running via "go run"
	// "go run" creates temporary executables in go-build cache directory
	if strings.Contains(exePath, "/go-build/") ||
		strings.Contains(exePath, "\\go-build\\") ||
		strings.Contains(exePath, filepath.Join(os.TempDir(), "go-build")) {
		fmt.Println("[Lokstra] Detected: go run (temporary build)")
		return RunModeDev
	}

	// 4️⃣ Additional check for go run on different systems
	// Check if executable is in system temp directory (common for go run)
	tempDir := filepath.ToSlash(os.TempDir())
	if strings.HasPrefix(exePath, tempDir) {
		fmt.Println("[Lokstra] Detected: go run (temp directory)")
		return RunModeDev
	}

	// 5️⃣ Check if binary name suggests it's a compiled production binary
	// Compiled binaries usually have specific names (not random hashes)
	// and are located in project directory or system paths
	if !strings.Contains(exeName, "exe") ||
		(strings.Contains(exeName, ".exe") && len(exeName) > 10) {
		// If we're in the project directory with a named binary
		wd, _ := os.Getwd()
		if strings.Contains(exePath, filepath.ToSlash(wd)) {
			fmt.Println("[Lokstra] Detected: compiled binary in project directory")
			return RunModeProd
		}
	}

	// 6️⃣ Default: assume production mode
	fmt.Println("[Lokstra] Detected: production binary (default)")
	return RunModeProd
}

// runAutoGen triggers the annotation scanner and route generator
func runAutoGen() (bool, error) {
	return annotation.ProcessComplexAnnotations("", 0,
		func(ctx *annotation.RouterServiceContext) error {
			fmt.Printf("Processing folder: %s\n", ctx.FolderPath)
			fmt.Printf("  - Skipped: %d files\n", len(ctx.SkippedFiles))
			fmt.Printf("  - Updated: %d files\n", len(ctx.UpdatedFiles))
			fmt.Printf("  - Deleted: %d files\n", len(ctx.DeletedFiles))

			// Generate code
			if err := annotation.GenerateCodeForFolder(ctx); err != nil {
				return err
			}

			return nil
		})
}

// deleteAllCacheFiles removes all zz_cache.lokstra.json files to force rebuild
func deleteAllCacheFiles() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	return filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}
		if !info.IsDir() && info.Name() == "zz_cache.lokstra.json" {
			fmt.Printf("  Deleting cache: %s\n", path)
			os.Remove(path)
		}
		return nil
	})
}

// relaunchWithGoRun restarts the current app using "go run ."
func relaunchWithGoRun() {
	fmt.Println("[Lokstra] Relaunching with go run...")

	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=1", childEnvKey))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Println("[Lokstra] Relaunch (go run) failed:", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// relaunchWithDlv handles debug mode - exits with message to restart debugger
func relaunchWithDlv() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  [Lokstra] AUTOGEN COMPLETED - DEBUGGER RESTART REQUIRED       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println("")
	fmt.Println("⚠️  Code generation detected changes.")
	fmt.Println("⚠️  Please STOP and RESTART your debugger to load the new code.")
	fmt.Println("")
	fmt.Println("Press Ctrl+C or stop the debugger, then press F5 to restart.")
	fmt.Println("")

	// Exit cleanly so debugger can be restarted
	os.Exit(0)
}

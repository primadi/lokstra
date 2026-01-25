package lokstra_init

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/annotation"
	"github.com/primadi/lokstra/lokstra_registry"
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
// scanPath specifies additional paths to scan for annotations (besides current working directory).
// If --generate-only flag is present, it will only run code generation and exit.
// Example:
//
//	func main() {
//	    lokstra.Bootstrap("./services", "./custom_modules")
//	    // ... rest of main ...
//	}
func Bootstrap(scanPath ...string) {
	// 1️⃣ Check for --generate-only flag (case-insensitive)
	generateOnly := false
	for _, arg := range os.Args {
		// Case-insensitive check: --generate-only, --GENERATE-ONLY, etc.
		if strings.ToLower(arg) == "--generate-only" {
			generateOnly = true
			break
		}
	}

	if generateOnly {
		fmt.Println("[Lokstra] Running in GENERATE-ONLY mode")
		fmt.Println("[Lokstra] Force rebuilding all generated code...")

		// Force rebuild by deleting all cache files first
		if err := deleteAllCacheFiles(scanPath); err != nil {
			fmt.Println("[Lokstra] Warning: failed to delete cache files:", err)
		}

		// Run autogen
		_, err := runAutoGen(scanPath)
		if err != nil {
			fmt.Println("[Lokstra] Autogen failed:", err)
			os.Exit(1)
		}

		fmt.Println("[Lokstra] ✅ Code generation completed successfully")
		os.Exit(0)
	}

	// 2️⃣ Detect mode and repository in config for runtime access
	Mode = DetectRunMode()
	lokstra_registry.SetConfig("runtime.mode", string(Mode))
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
	codeChanged, err := runAutoGen(scanPath)
	if err != nil {
		fmt.Println("[Lokstra] Autogen failed:", err)
		os.Exit(1)
	}

	// 6️⃣ Relaunch using correct method (only if code changed)
	if codeChanged {
		switch Mode {
		case RunModeDebug:
			relaunchWithDlv()
		case RunModeDev:
			relaunchWithGoRun()
		default:
			fmt.Println("[Lokstra] Unknown mode — continuing normally.")
		}
	}
}

// DetectRunMode inspects runtime and env to detect how the app was started.
func DetectRunMode() RunMode {
	exe, err := os.Executable()
	if err != nil {
		logger.LogDebug("[Lokstra] Warning: cannot get executable path:", err)
		return RunModeProd
	}

	exePath := filepath.ToSlash(exe)
	exeName := filepath.Base(exe)
	logger.LogDebug("[Lokstra] Executable: %s\n", exePath)

	// 1️⃣ Check if running under Delve debugger
	// Delve wraps the binary with __debug_bin
	if strings.Contains(exeName, "__debug_bin") {
		logger.LogDebug("[Lokstra] Detected: Delve debugger (debug binary)")
		return RunModeDebug
	}

	// 2️⃣ Check if running via "go run"
	// "go run" creates temporary executables in go-build cache directory
	if strings.Contains(exePath, "go-build") ||
		strings.Contains(exePath, os.TempDir()) {
		logger.LogDebug("[Lokstra] Detected: go run (temporary build)")
		return RunModeDev
	}

	// 3️⃣ If we reach here, it's a compiled binary (not go run, not debugger)
	// Windows: .exe extension confirms it's a compiled binary
	// Linux/Mac: no .exe, but also not in temp/go-build, so it's compiled
	// Default to production mode for all compiled binaries
	logger.LogDebug("[Lokstra] Detected: compiled binary (production mode)")
	return RunModeProd
}

// runAutoGen triggers the annotation scanner and route generator.
// It scans the current working directory plus any additional paths provided in scanPath.
func runAutoGen(scanPath []string) (bool, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Build the list of paths to scan: current directory + scanPath
	pathsToScan := []string{wd}
	pathsToScan = append(pathsToScan, scanPath...)

	return annotation.ProcessComplexAnnotations(pathsToScan, 0,
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

// deleteAllCacheFiles removes all zz_cache.lokstra.json files to force rebuild.
// It scans the current working directory plus any additional paths provided in scanPath.
func deleteAllCacheFiles(scanPath []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Build the list of paths to scan: current directory + scanPath
	pathsToScan := []string{wd}
	pathsToScan = append(pathsToScan, scanPath...)

	// Walk through all paths
	for _, scanDir := range pathsToScan {
		err := filepath.Walk(scanDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors, continue walking
			}
			if !info.IsDir() && info.Name() == "zz_cache.lokstra.json" {
				fmt.Printf("  Deleting cache: %s\n", path)
				os.Remove(path)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("  Warning: failed to walk directory %s: %v\n", scanDir, err)
		}
	}

	return nil
}

// relaunchWithGoRun restarts the current app using "go run ."
func relaunchWithGoRun() {
	fmt.Println("[Lokstra] Relaunching with go run...")

	// Build command args: "go run ." + original program args (skip os.Args[0])
	args := []string{"run", "."}
	if len(os.Args) > 1 {
		args = append(args, os.Args[1:]...)
	}

	cmd := exec.Command("go", args...)
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
	fmt.Println("⚠️  Please RESTART your debugger to load the new code.")
	fmt.Println("")

	// Exit cleanly so debugger can be restarted
	os.Exit(0)
}

// return runtime mode: dev, debug, or prod
func GetRuntimeMode() string {
	return lokstra_registry.GetRuntimeMode()
}

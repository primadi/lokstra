package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "new":
		newCmd()
	case "autogen":
		autogenCmd()
	case "version":
		fmt.Printf("Lokstra CLI v%s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Lokstra CLI - Create new Lokstra projects from templates")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  lokstra new <project-name> [flags]")
	fmt.Println("  lokstra autogen [folder]")
	fmt.Println("  lokstra version")
	fmt.Println("  lokstra help")
	fmt.Println()
	fmt.Println("Flags for 'new' command:")
	fmt.Println("  -template <name>    Template to use (optional, interactive if not specified)")
	fmt.Println("  -branch <name>      Git branch to download from (default: dev2)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  lokstra new myapp")
	fmt.Println("  lokstra new myapp -template 02_app_framework/01_medium_system")
	fmt.Println("  lokstra new myapp -template 01_router/01_router_only -branch main")
	fmt.Println()
	fmt.Println("  lokstra autogen                 # Generate code in current directory")
	fmt.Println("  lokstra autogen ./myproject     # Generate code in specific folder")
}

func newCmd() {
	// Parse flags for 'new' command
	newFlags := flag.NewFlagSet("new", flag.ExitOnError)
	templateFlag := newFlags.String("template", "", "Template to use")
	branchFlag := newFlags.String("branch", "dev2", "Git branch to download from")

	// Get project name (first argument after 'new')
	if len(os.Args) < 3 {
		fmt.Println("Error: project name is required")
		fmt.Println()
		fmt.Println("Usage: lokstra new <project-name> [flags]")
		os.Exit(1)
	}

	projectName := os.Args[2]

	// Parse remaining flags
	newFlags.Parse(os.Args[3:])

	// Execute new command
	if err := executeNew(projectName, *templateFlag, *branchFlag); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func executeNew(projectName, templatePath, branch string) error {
	fmt.Printf("🚀 Creating new Lokstra project: %s\n\n", projectName)

	// If template not specified, show interactive selection
	if templatePath == "" {
		var err error
		templatePath, err = selectTemplate(branch)
		if err != nil {
			return err
		}
	}

	fmt.Printf("📦 Selected template: %s\n", templatePath)
	fmt.Printf("🌿 Branch: %s\n\n", branch)

	// Execute the creation process
	creator := NewProjectCreator(projectName, templatePath, branch)
	return creator.Create()
}

func autogenCmd() {
	// Get target folder (optional, defaults to current directory)
	targetFolder := "."
	if len(os.Args) >= 3 {
		targetFolder = os.Args[2]
	}

	// Execute autogen
	if err := executeAutogen(targetFolder); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func executeAutogen(targetFolder string) error {
	fmt.Printf("🔧 Running code generation in: %s\n\n", targetFolder)

	// Check if target folder exists
	if _, err := os.Stat(targetFolder); os.IsNotExist(err) {
		return fmt.Errorf("folder does not exist: %s", targetFolder)
	}

	// Run: go run . --generate-only
	return runCommandInDir(targetFolder, "go", "run", ".", "--generate-only")
}

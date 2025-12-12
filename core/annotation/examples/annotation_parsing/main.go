package main

import (
	"fmt"
	"strings"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/annotation"
)

func main() {
	filePath := "annotation_example.go"

	fmt.Println("Parsing file:", filePath)
	fmt.Println(strings.Repeat("=", 70))

	annotations, err := annotation.ParseFileAnnotations(filePath)
	if err != nil {
		logger.LogPanic("Error parsing file: %v", err)
	}

	fmt.Printf("\nFound %d annotations:\n\n", len(annotations))

	for i, ann := range annotations {
		fmt.Printf("%d. @%s (line %d) -> target: %s\n", i+1, ann.Name, ann.Line, ann.TargetName)

		// Show args if any
		if len(ann.Args) > 0 {
			fmt.Println("   Args:")
			for k, v := range ann.Args {
				fmt.Printf("     %s = %v\n", k, v)
			}
		}
		fmt.Println()
	}

	// Expected: Only annotations from UserService (lines 27, 30, 34, 38)
	// Should NOT include: RegisterMiddleware example (line 13)
	// Should NOT include: AnotherFunction example (line 50)

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\nExpected annotations:")
	fmt.Println("  1. @RouterService on UserService (line 27)")
	fmt.Println("  2. @Inject on UserRepo field (line 30)")
	fmt.Println("  3. @Route on GetByID method (line 34)")
	fmt.Println("  4. @Route on Create method (line 38)")
	fmt.Println("\nIgnored (TAB/multi-space indented):")
	fmt.Println("  - Line 13: Documentation example (TAB-indented)")
	fmt.Println("  - Line 50: Documentation example (multi-space indented)")
}

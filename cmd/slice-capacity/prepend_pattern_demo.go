package main

import "fmt"

func Zmain() {
	fmt.Print("=== Pattern: Prepend with Slice Literal ===\n")

	// Original slice
	original := []string{"app1", "app2", "app3"}
	fmt.Printf("Before: %v (len=%d, cap=%d)\n", original, len(original), cap(original))

	// PREPEND using slice literal (like in builder.go)
	newApp := "newApp"
	original = append([]string{newApp}, original...)

	fmt.Printf("After:  %v (len=%d, cap=%d)\n", original, len(original), cap(original))
	fmt.Println("\n✅ AMAN karena:")
	fmt.Println("   1. []string{newApp} adalah LITERAL slice baru (len=1, cap=1)")
	fmt.Println("   2. append() PASTI allocate new array (cap tidak cukup)")
	fmt.Println("   3. Result di-assign kembali ke variable yang sama")

	fmt.Print("\n=== Simulasi Multiple Calls ===\n")

	apps := []string{"app1", "app2"}
	fmt.Printf("Initial: %v\n", apps)

	// Prepend multiple times
	apps = append([]string{"new1"}, apps...)
	fmt.Printf("After prepend 'new1': %v\n", apps)

	apps = append([]string{"new2"}, apps...)
	fmt.Printf("After prepend 'new2': %v\n", apps)

	apps = append([]string{"new3"}, apps...)
	fmt.Printf("After prepend 'new3': %v\n", apps)

	fmt.Println("\n✅ Semua prepend berhasil tanpa aliasing!")

	fmt.Print("\n=== Comparison: UNSAFE Pattern ===\n")

	base := []string{"a", "b", "c", "d", "e"}
	fmt.Printf("base: %v (len=%d, cap=%d)\n", base, len(base), cap(base))

	// ❌ UNSAFE: append to different variable
	result1 := append(base, "x")
	result2 := append(base, "y")

	fmt.Printf("result1: %v (len=%d, cap=%d)\n", result1, len(result1), cap(result1))
	fmt.Printf("result2: %v (len=%d, cap=%d)\n", result2, len(result2), cap(result2))
	fmt.Printf("❌ result1[5] = %s (overwritten!)\n", result1[5])

	fmt.Println("\n=== Summary ===")
	fmt.Println("✅ AMAN:")
	fmt.Println("   x = append([]T{item}, x...)  // Literal slice, always reallocate")
	fmt.Println("   x = append(x, item)          // Append to self")
	fmt.Println("")
	fmt.Println("❌ BAHAYA:")
	fmt.Println("   y := append(x, item)         // Different variable, might alias")
}

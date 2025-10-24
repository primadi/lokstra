package main

import "fmt"

func Ymain() {
	fmt.Println("=== BAHAYA: x := append(source, ...) ===")
	fmt.Print("Source dan dest BERBEDA → bisa share underlying array\n")

	source := make([]string, 0)
	for i := 0; i < 5; i++ {
		source = append(source, fmt.Sprintf("item%d", i))
	}
	fmt.Printf("source: len=%d, cap=%d\n", len(source), cap(source))

	// ❌ BAHAYA: append ke variable BARU
	dest1 := append(source, "route1")
	dest2 := append(source, "route2")

	fmt.Printf("dest1: len=%d, cap=%d, last=%s\n", len(dest1), cap(dest1), dest1[len(dest1)-1])
	fmt.Printf("dest2: len=%d, cap=%d, last=%s\n", len(dest2), cap(dest2), dest2[len(dest2)-1])
	fmt.Printf("❌ dest1[5] = %s (OVERWRITTEN by dest2!)\n", dest1[5])
	fmt.Printf("❌ dest2[5] = %s\n\n", dest2[5])

	fmt.Println("=== AMAN: x = append(x, ...) ===")
	fmt.Print("Source dan dest SAMA → tidak masalah share array\n")

	items := make([]string, 0)
	for i := 0; i < 5; i++ {
		items = append(items, fmt.Sprintf("item%d", i))
	}
	fmt.Printf("Before: len=%d, cap=%d\n", len(items), cap(items))

	// ✅ AMAN: append ke variable SAMA
	items = append(items, "new1")
	fmt.Printf("After append 'new1': len=%d, cap=%d, last=%s\n", len(items), cap(items), items[len(items)-1])

	items = append(items, "new2")
	fmt.Printf("After append 'new2': len=%d, cap=%d, last=%s\n", len(items), cap(items), items[len(items)-1])

	fmt.Printf("✅ items[5] = %s\n", items[5])
	fmt.Printf("✅ items[6] = %s\n\n", items[6])

	fmt.Println("=== GOLDEN RULES ===")
	fmt.Println("❌ BERBAHAYA:")
	fmt.Println("   newSlice := append(oldSlice, x)  // Bisa alias!")
	fmt.Println("")
	fmt.Println("✅ AMAN (Option 1 - append to self):")
	fmt.Println("   slice = append(slice, x)  // Grow the same slice")
	fmt.Println("")
	fmt.Println("✅ AMAN (Option 2 - explicit copy):")
	fmt.Println("   newSlice := make([]T, len(oldSlice)+1)")
	fmt.Println("   copy(newSlice, oldSlice)")
	fmt.Println("   newSlice[len(oldSlice)] = x")
}

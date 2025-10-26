package main

import "fmt"

type Container struct {
	Items *[]string
}

func Amain() {
	fmt.Println("=== Test 1: Append to nil pointer slice ===")

	c := &Container{}
	fmt.Printf("Before: c.Items == nil? %v\n", c.Items == nil)

	// This will panic because we're dereferencing nil pointer
	// *c.Items = append(*c.Items, "item1")

	fmt.Println("\n=== Test 2: Initialize first, then append ===")
	c2 := &Container{}
	items := []string(nil) // nil slice
	c2.Items = &items

	fmt.Printf("Before: c2.Items == nil? %v\n", c2.Items == nil)
	fmt.Printf("Before: *c2.Items == nil? %v\n", *c2.Items == nil)
	fmt.Printf("Before: len(*c2.Items) = %d\n", len(*c2.Items))

	// ✅ This works because *c2.Items is a nil slice (not nil pointer)
	*c2.Items = append(*c2.Items, "item1")
	fmt.Printf("After append: %v\n", *c2.Items)

	*c2.Items = append(*c2.Items, "item2")
	fmt.Printf("After append: %v\n", *c2.Items)

	fmt.Println("\n=== Test 3: Direct nil slice (without pointer) ===")
	var direct []string // nil slice
	fmt.Printf("direct == nil? %v\n", direct == nil)

	direct = append(direct, "a")
	fmt.Printf("After append: %v (len=%d, cap=%d)\n", direct, len(direct), cap(direct))

	direct = append(direct, "b")
	fmt.Printf("After append: %v (len=%d, cap=%d)\n", direct, len(direct), cap(direct))

	fmt.Println("\n=== Conclusion ===")
	fmt.Println("✅ append(nil_slice, x) is SAFE in Go")
	fmt.Println("✅ *pointer_to_nil_slice = append(*pointer_to_nil_slice, x) is SAFE")
	fmt.Println("❌ BUT: pointer itself must NOT be nil (must point to nil slice)")
}

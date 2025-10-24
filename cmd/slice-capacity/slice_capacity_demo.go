package main

import "fmt"

func Xmain() {
	fmt.Println("=== Go Slice Capacity Growth ===")
	var s []int

	for i := 1; i <= 10; i++ {
		s = append(s, i)
		fmt.Printf("len=%d, cap=%d\n", len(s), cap(s))
	}

	fmt.Println("\n=== Simulasi Bug Middleware ===")

	// Simulasi: kita punya middleware slice dengan berbagai jumlah
	for mwCount := 1; mwCount <= 8; mwCount++ {
		mw := make([]string, 0)

		// Isi middleware
		for i := 0; i < mwCount; i++ {
			mw = append(mw, fmt.Sprintf("mw%d", i))
		}

		fmt.Printf("\n%d middleware: len=%d, cap=%d\n", mwCount, len(mw), cap(mw))

		// Simulasi append handler (seperti di NewHandler)
		handlers1 := append(mw, "handler1")
		handlers2 := append(mw, "handler2")

		fmt.Printf("  handlers1: len=%d, cap=%d, last=%s\n",
			len(handlers1), cap(handlers1), handlers1[len(handlers1)-1])
		fmt.Printf("  handlers2: len=%d, cap=%d, last=%s\n",
			len(handlers2), cap(handlers2), handlers2[len(handlers2)-1])

		// Check apakah mereka share underlying array
		if len(mw) < cap(mw) {
			fmt.Printf("  ⚠️  BAHAYA! len < cap, append akan REUSE array\n")
			fmt.Printf("  handlers1[%d] = %s (should be handler1, but might be overwritten)\n",
				len(mw), handlers1[len(mw)])
		} else {
			fmt.Printf("  ✅ AMAN! len == cap, append akan allocate NEW array\n")
		}
	}

	fmt.Println("\n=== Detailed Analysis: Why 5 is dangerous ===")
	mw := make([]string, 0)
	for i := 0; i < 5; i++ {
		mw = append(mw, fmt.Sprintf("mw%d", i))
	}
	fmt.Printf("After adding 5 middleware: len=%d, cap=%d\n", len(mw), cap(mw))
	fmt.Println("When we do: handlers1 := append(mw, handler1)")
	fmt.Println("           handlers2 := append(mw, handler2)")
	fmt.Println("Both handlers1 and handlers2 will REUSE the SAME underlying array!")
	fmt.Println("Because len(5) < cap(6), so append doesn't allocate new memory")
}

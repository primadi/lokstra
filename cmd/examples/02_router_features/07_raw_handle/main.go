package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/router"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	ctx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(ctx, "test-app", ":8080")

	// Test 1: Direct file server
	app.RawHandle("/files/", true, http.FileServer(http.Dir("./files")))

	// Test 2: StaticFallback with embed.FS and fs.FS (complete fallback chain)
	// Create separate fs.FS from disk directory
	fsdataFS := os.DirFS("./fsdata")

	fallback := router.NewStaticFallback(
		os.DirFS("./priority01"), // 1st priority: disk custom files
		os.DirFS("./default"),    // 2nd priority: disk default files
	).
		WithEmbedFS(staticFiles, "static").
		WithSourceFS(fsdataFS)

	app.RawHandle("/assets/", true, fallback.RawHandler(false))

	// Test 3: Direct embed.FS server
	app.RawHandle("/embed/", true, http.FileServer(http.FS(staticFiles)))

	// Simple info endpoint
	app.GET("/", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"message":     "RawHandle Test with Complete Fallback Chain",
			"description": "Demonstrates all NewStaticFallback source types",
			"fallback_priority": []string{
				"1. ./priority01/* (http.Dir - disk)",
				"2. ./default/* (http.Dir - disk)",
				"3. embed.FS static/* (embed.FS - auto-subFirstDir)",
				"4. ./fsdata/* (fs.FS from os.DirFS - no subFirstDir)",
			},
			"note":          "fs.FS now uses separate ./fsdata directory, completely different from embed.FS",
			"test_strategy": "Remove files step by step to see fallback in action",
			"endpoints": []string{
				"/files/sample.txt",
				"/assets/app.js (test fallback behavior)",
				"/embed/static/app.js (direct embed.FS)",
			},
		})
	})

	fmt.Println("RawHandle Test with Complete Fallback Chain")
	fmt.Println("Fallback priority for /assets/*:")
	fmt.Println("  1. ./priority01/* (http.Dir - disk)")
	fmt.Println("  2. ./default/* (http.Dir - disk)")
	fmt.Println("  3. embed.FS static/* (auto-subFirstDir)")
	fmt.Println("  4. ./fsdata/* (fs.FS from os.DirFS - separate source)")
	fmt.Println("")
	fmt.Println("Testing fallback behavior:")
	fmt.Println("  - Start with all sources available")
	fmt.Println("  - Remove ./priority01/app.js → fallback to ./default/")
	fmt.Println("  - Remove ./default/app.js → fallback to embed.FS")
	fmt.Println("  - Remove embed file → fallback to fs.FS (./fsdata/)")
	fmt.Println("  - Test fs-only.js (only exists in ./fsdata/)")
	fmt.Println("")
	fmt.Println("Test server starting on :8080")
	fmt.Println("Test URLs:")
	fmt.Println("  http://localhost:8080/")
	fmt.Println("  http://localhost:8080/files/sample.txt")
	fmt.Println("  http://localhost:8080/assets/app.js")
	fmt.Println("  http://localhost:8080/embed/static/app.js")

	if err := app.Start(); err != nil {
		panic(err)
	}
}

package main

import (
	"fmt"
	"strings"

	"github.com/primadi/lokstra/cmd/examples-old/01_basic_overview/05_client_rpc/hello_service"
)

func printSeparator(title string) {
	fmt.Printf("\n%s %s %s\n", strings.Repeat("=", 20), title, strings.Repeat("=", 20))
}

func printSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

func printError(operation string, err error) {
	fmt.Printf("❌ %s Error: %v\n", operation, err)
}

func demoStringReturn(client *hello_service.GreetingServiceClient) {
	printSeparator("STRING RETURN TYPE")

	testNames := []string{"World", "Lokstra", "Go Developer", ""}
	for _, name := range testNames {
		if greeting, err := client.Hello(name); err != nil {
			printError(fmt.Sprintf("Hello(%q)", name), err)
		} else {
			printSuccess(fmt.Sprintf("Hello(%q) → %s", name, greeting))
		}
	}
}

func demoInterfaceReturn(client *hello_service.GreetingServiceClient) {
	printSeparator("INTERFACE RETURN TYPE")

	testIDs := []int{123, 456, 0, -1}
	for _, id := range testIDs {
		if user, err := client.GetUser(id); err != nil {
			printError(fmt.Sprintf("GetUser(%d)", id), err)
		} else {
			printSuccess(fmt.Sprintf("GetUser(%d) → User{ID:%d, Name:%s, Email:%s, Active:%t}",
				id, user.GetID(), user.GetName(), user.GetEmail(), user.IsActive()))
		}
	}
}

func demoSliceInterfaceReturn(client *hello_service.GreetingServiceClient) {
	printSeparator("SLICE INTERFACE RETURN TYPE")

	testLimits := []int{3, 5, 0, 101}
	for _, limit := range testLimits {
		if users, err := client.GetUsers(limit); err != nil {
			printError(fmt.Sprintf("GetUsers(%d)", limit), err)
		} else {
			printSuccess(fmt.Sprintf("GetUsers(%d) → %d users", limit, len(users)))
			for i, user := range users {
				if i < 3 { // Show first 3 users
					fmt.Printf("   [%d] ID:%d, Name:%s, Active:%t\n",
						i, user.GetID(), user.GetName(), user.IsActive())
				} else if i == 3 {
					fmt.Printf("   ... and %d more users\n", len(users)-3)
					break
				}
			}
		}
	}
}

func demoMapReturn(client *hello_service.GreetingServiceClient) {
	printSeparator("MAP RETURN TYPE")

	testIDs := []int{123, 456, 0}
	for _, id := range testIDs {
		if stats, err := client.GetUserStats(id); err != nil {
			printError(fmt.Sprintf("GetUserStats(%d)", id), err)
		} else {
			printSuccess(fmt.Sprintf("GetUserStats(%d) → %d fields", id, len(stats)))
			for key, value := range stats {
				fmt.Printf("   %s: %v (type: %T)\n", key, value, value)
			}
		}
	}
}

func demoStructReturn(client *hello_service.GreetingServiceClient) {
	printSeparator("STRUCT RETURN TYPE")

	if info, err := client.GetSystemInfo(); err != nil {
		printError("GetSystemInfo", err)
	} else {
		printSuccess("GetSystemInfo → SystemInfo")
		fmt.Printf("   Version: %s\n", info.Version)
		fmt.Printf("   Uptime: %s\n", info.Uptime)
		fmt.Printf("   Memory: %s\n", info.Memory)
		fmt.Printf("   CPU Usage: %.1f%%\n", info.CPUUsage)
		fmt.Printf("   Connected: %d\n", info.Connected)
	}
}

func demoPrimitiveReturns(client *hello_service.GreetingServiceClient) {
	printSeparator("PRIMITIVE RETURN TYPES")

	// Test int return
	if count, err := client.GetUserCount(); err != nil {
		printError("GetUserCount", err)
	} else {
		printSuccess(fmt.Sprintf("GetUserCount → %d", count))
	}

	// Test bool return
	testIDs := []int{124, 125, 0}
	for _, id := range testIDs {
		if active, err := client.GetUserActive(id); err != nil {
			printError(fmt.Sprintf("GetUserActive(%d)", id), err)
		} else {
			printSuccess(fmt.Sprintf("GetUserActive(%d) → %t", id, active))
		}
	}

	// Test time return
	if serverTime, err := client.GetServerTime(); err != nil {
		printError("GetServerTime", err)
	} else {
		printSuccess(fmt.Sprintf("GetServerTime → %s", serverTime.Format("2006-01-02 15:04:05")))
	}
}

func demoDynamicReturn(client *hello_service.GreetingServiceClient) {
	printSeparator("DYNAMIC any RETURN TYPE")

	testTypes := []string{"user", "stats", "message", "number", "list", "unknown"}
	for _, dataType := range testTypes {
		if data, err := client.GetDynamicData(dataType); err != nil {
			printError(fmt.Sprintf("GetDynamicData(%s)", dataType), err)
		} else {
			printSuccess(fmt.Sprintf("GetDynamicData(%s) → %T", dataType, data))
			switch v := data.(type) {
			case map[string]any:
				fmt.Printf("   Map with %d keys\n", len(v))
			case []any:
				fmt.Printf("   Slice with %d items\n", len(v))
			case string:
				fmt.Printf("   String: %q\n", v)
			default:
				fmt.Printf("   Value: %v\n", v)
			}
		}
	}
}

func demoVoidOperations(client *hello_service.GreetingServiceClient) {
	printSeparator("VOID OPERATIONS (ERROR ONLY)")

	// Test Ping
	if err := client.Ping(); err != nil {
		printError("Ping", err)
	} else {
		printSuccess("Ping → Success")
	}

	// Test ClearCache
	if err := client.ClearCache(); err != nil {
		printError("ClearCache", err)
	} else {
		printSuccess("ClearCache → Success")
	}

	// Test DeleteUser
	testIDs := []int{999, 0}
	for _, id := range testIDs {
		if err := client.DeleteUser(id); err != nil {
			printError(fmt.Sprintf("DeleteUser(%d)", id), err)
		} else {
			printSuccess(fmt.Sprintf("DeleteUser(%d) → Success", id))
		}
	}
}

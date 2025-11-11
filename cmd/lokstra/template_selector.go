package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Template represents a project template
type Template struct {
	Path         string   `json:"path"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Complexity   int      `json:"complexity"`
	Tags         []string `json:"tags"`
	MinGoVersion string   `json:"minGoVersion"`
}

// TemplateList represents the templates.json structure
type TemplateList struct {
	Version   string     `json:"version"`
	Templates []Template `json:"templates"`
}

// fetchTemplates downloads the template list from GitHub
func fetchTemplates(branch string) ([]Template, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/primadi/lokstra/%s/project_templates/templates.json", branch)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download templates list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download templates list: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates list: %w", err)
	}

	var templateList TemplateList
	if err := json.Unmarshal(body, &templateList); err != nil {
		return nil, fmt.Errorf("failed to parse templates list: %w", err)
	}

	return templateList.Templates, nil
}

func selectTemplate(branch string) (string, error) {
	// Download template list from GitHub
	fmt.Println("📥 Fetching available templates...")
	templates, err := fetchTemplates(branch)
	if err != nil {
		return "", err
	}
	fmt.Println()
	fmt.Println("📋 Available Templates:")
	fmt.Println()

	// Group by category
	currentCategory := ""
	for i, tmpl := range templates {
		if tmpl.Category != currentCategory {
			currentCategory = tmpl.Category
			fmt.Printf("\n%s:\n", currentCategory)
		}
		fmt.Printf("  [%d] %s\n", i+1, tmpl.Name)
		fmt.Printf("      %s\n", tmpl.Description)
	}

	fmt.Println()
	fmt.Print("Select template (1-", len(templates), "): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(templates) {
		return "", fmt.Errorf("invalid selection: %s", input)
	}

	selected := templates[selection-1]
	fmt.Printf("\n✅ Selected: %s\n", selected.Name)
	fmt.Println()

	return selected.Path, nil
}

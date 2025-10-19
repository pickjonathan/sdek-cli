package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pickjonathan/sdek-cli/internal/report"
	"github.com/spf13/cobra"
)

var (
	htmlInputFile  string
	htmlOutputFile string
)

// htmlCmd represents the html command
var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "Generate an interactive HTML compliance report",
	Long: `Generate an interactive HTML dashboard from a JSON report file.

The HTML report provides:
- Visual compliance dashboard with charts and gauges
- Interactive framework and control exploration
- Filterable evidence with AI enhancement indicators
- Detailed findings analysis with severity indicators
- Expandable control details with full context

Example usage:
  # Generate HTML from default report location
  sdek html

  # Specify input and output files
  sdek html --input ~/sdek-report.json --output ~/compliance-dashboard.html

  # Use short flags
  sdek html -i report.json -o dashboard.html`,
	RunE: runHTML,
}

func init() {
	rootCmd.AddCommand(htmlCmd)

	// Define flags
	htmlCmd.Flags().StringVarP(&htmlInputFile, "input", "i", "", "Input JSON report file (default: ~/sdek-report.json)")
	htmlCmd.Flags().StringVarP(&htmlOutputFile, "output", "o", "", "Output HTML file (default: ~/sdek-report.html)")
}

func runHTML(cmd *cobra.Command, args []string) error {
	// Determine input file
	inputPath := htmlInputFile
	if inputPath == "" {
		// Default to ~/sdek-report.json
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		inputPath = filepath.Join(homeDir, "sdek-report.json")
	}

	// Expand ~ in path
	if strings.HasPrefix(inputPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		inputPath = filepath.Join(homeDir, inputPath[1:])
	}

	// Check if input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s\n\nPlease run 'sdek report' first to generate a JSON report", inputPath)
	}

	// Determine output file
	outputPath := htmlOutputFile
	if outputPath == "" {
		// Default based on input file name
		if htmlInputFile == "" {
			// User didn't specify input, use default output
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			outputPath = filepath.Join(homeDir, "sdek-report.html")
		} else {
			// Use input filename with .html extension
			ext := filepath.Ext(inputPath)
			outputPath = strings.TrimSuffix(inputPath, ext) + ".html"
		}
	}

	// Expand ~ in output path
	if strings.HasPrefix(outputPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputPath = filepath.Join(homeDir, outputPath[1:])
	}

	fmt.Printf("📊 Generating HTML report...\n")
	fmt.Printf("   Input:  %s\n", inputPath)
	fmt.Printf("   Output: %s\n\n", outputPath)

	// Generate HTML
	if err := report.GenerateHTML(inputPath, outputPath); err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	// Get file size
	info, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to stat output file: %w", err)
	}

	sizeKB := float64(info.Size()) / 1024
	var sizeStr string
	if sizeKB > 1024 {
		sizeStr = fmt.Sprintf("%.1f MB", sizeKB/1024)
	} else {
		sizeStr = fmt.Sprintf("%.1f KB", sizeKB)
	}

	fmt.Printf("✅ HTML report generated successfully!\n\n")
	fmt.Printf("📁 File: %s (%s)\n", outputPath, sizeStr)
	fmt.Printf("🌐 Open in browser: file://%s\n\n", outputPath)
	fmt.Printf("💡 Tip: The HTML file is self-contained and can be:\n")
	fmt.Printf("   • Opened directly in any web browser\n")
	fmt.Printf("   • Shared with stakeholders\n")
	fmt.Printf("   • Archived for compliance audits\n")

	return nil
}

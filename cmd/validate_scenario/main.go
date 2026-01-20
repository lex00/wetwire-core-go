// validate_scenario validates scenario results against defined validation rules
// and expected files.
//
// Usage:
//
//	go run ./cmd/validate_scenario [scenario_path] [results_dir] [persona] [flags]
//
// Flags:
//
//	--markdown  Output report in markdown format
//	--json      Output report in JSON format
//	--quiet     Only output pass/fail status
//
// Examples:
//
//	go run ./cmd/validate_scenario ./examples/honeycomb_k8s
//	go run ./cmd/validate_scenario ./examples/honeycomb_k8s ./results/intermediate
//	go run ./cmd/validate_scenario ./examples/honeycomb_k8s intermediate
//	go run ./cmd/validate_scenario ./examples/honeycomb_k8s --markdown > VALIDATION.md
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-core-go/scenario"
	"github.com/lex00/wetwire-core-go/scenario/validator"
	"gopkg.in/yaml.v3"
)

func main() {
	// Parse arguments
	scenarioPath := ""
	resultsDir := ""
	persona := "intermediate"
	outputMarkdown := false
	outputJSON := false
	quiet := false

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--markdown", "-m":
			outputMarkdown = true
		case "--json", "-j":
			outputJSON = true
		case "--quiet", "-q":
			quiet = true
		case "--help", "-h":
			printUsage()
			return
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", arg)
				os.Exit(1)
			}
			if scenarioPath == "" {
				scenarioPath = arg
			} else if resultsDir == "" {
				// Could be a results dir or a persona name
				if isPersonaName(arg) {
					persona = arg
				} else {
					resultsDir = arg
				}
			} else {
				// Third positional arg must be persona
				persona = arg
			}
		}
	}

	if scenarioPath == "" {
		fmt.Fprintln(os.Stderr, "Error: scenario path is required")
		printUsage()
		os.Exit(1)
	}

	// Resolve absolute path
	absScenarioPath, err := filepath.Abs(scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving scenario path: %v\n", err)
		os.Exit(1)
	}

	// Default results dir
	if resultsDir == "" {
		resultsDir = filepath.Join(absScenarioPath, "results", persona)
	} else if !filepath.IsAbs(resultsDir) {
		// Check if it's a relative path from scenario dir or from cwd
		if info, err := os.Stat(resultsDir); err == nil && info.IsDir() {
			resultsDir, _ = filepath.Abs(resultsDir)
		} else {
			// Try as relative to scenario path
			resultsDir = filepath.Join(absScenarioPath, resultsDir)
		}
	}

	// Load scenario config
	scenarioConfig, err := loadScenarioConfig(absScenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading scenario config: %v\n", err)
		os.Exit(1)
	}

	if !quiet {
		fmt.Println("╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║           Wetwire Scenario Validator                       ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("Scenario: %s\n", scenarioConfig.Name)
		fmt.Printf("Results:  %s\n", resultsDir)
		fmt.Println()
	}

	// Check results dir exists
	if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: results directory does not exist: %s\n", resultsDir)
		os.Exit(1)
	}

	// Create validator and run
	v := validator.New(scenarioConfig, absScenarioPath, resultsDir)
	report, err := v.Validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during validation: %v\n", err)
		os.Exit(1)
	}

	// Output report
	if outputJSON {
		data, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(data))
	} else if outputMarkdown {
		fmt.Print(validator.FormatReportMarkdown(report))
	} else if !quiet {
		fmt.Print(validator.FormatReport(report))
	} else {
		// Quiet mode - just status
		if report.Passed {
			fmt.Println("PASSED")
		} else {
			fmt.Println("FAILED")
		}
	}

	// Exit with appropriate code
	if !report.Passed {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: validate_scenario [scenario_path] [results_dir|persona] [persona] [flags]

Validates scenario results against defined validation rules and expected files.

Arguments:
  scenario_path  Path to the scenario directory (required)
  results_dir    Path to results directory (default: scenario_path/results/persona)
  persona        Persona name to validate (default: intermediate)

Flags:
  --markdown, -m  Output report in markdown format
  --json, -j      Output report in JSON format
  --quiet, -q     Only output pass/fail status
  --help, -h      Show this help

Examples:
  validate_scenario ./examples/honeycomb_k8s
  validate_scenario ./examples/honeycomb_k8s intermediate
  validate_scenario ./examples/honeycomb_k8s ./results/intermediate
  validate_scenario ./examples/honeycomb_k8s --markdown > VALIDATION.md
  validate_scenario ./examples/honeycomb_k8s --json`)
}

func isPersonaName(s string) bool {
	// Check built-in personas
	builtIn := []string{"beginner", "intermediate", "expert"}
	for _, p := range builtIn {
		if s == p {
			return true
		}
	}
	// Custom personas are also valid - they don't start with special chars
	return !strings.HasPrefix(s, "-") && !strings.HasPrefix(s, ".") && !strings.HasPrefix(s, "/")
}

func loadScenarioConfig(scenarioDir string) (*scenario.ScenarioConfig, error) {
	configPath := filepath.Join(scenarioDir, "scenario.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading scenario.yaml: %w", err)
	}

	var config scenario.ScenarioConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing scenario.yaml: %w", err)
	}

	return &config, nil
}

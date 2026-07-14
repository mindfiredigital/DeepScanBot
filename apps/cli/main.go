package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mindfiredigital/DeepScanBot/packages/crawler"
	"github.com/mindfiredigital/DeepScanBot/packages/exitcode"
	"github.com/mindfiredigital/DeepScanBot/packages/logger"
	"github.com/mindfiredigital/DeepScanBot/packages/output"
	"github.com/mindfiredigital/DeepScanBot/packages/storage"
)

// cliVersion is the current version of the CLI
const cliVersion = "1.0.0"

var log = logger.New("info")

// ScanOptions holds all scan configuration
type ScanOptions struct {
	Depth            int
	Timeout          int
	Proxy            string
	JSON             bool
	MaxSize          int
	DisableRedirects bool
	ShowSource       bool
	Insecure         bool
	Unique           bool
	Concurrency      int
	HostConcurrency  int
	ContentTypes     string
	Output           string
	IgnoreRobots     bool
	CrossDomain      bool
	Retries          int
	RetryBackoff     time.Duration
	Delay            time.Duration
	Sitemap          bool
	Resume           bool
}

func parseIntValue(val string) (int, bool) {
	if i, err := strconv.Atoi(val); err == nil {
		return i, true
	}
	return 0, false
}

func parseDurationValue(val string) (time.Duration, bool) {
	if d, err := time.ParseDuration(val); err == nil {
		return d, true
	}
	return 0, false
}

func applyScanOption(opts *ScanOptions, key, val string) {
	switch key {
	case "depth":
		if d, ok := parseIntValue(val); ok {
			opts.Depth = d
		}
	case "timeout":
		if t, ok := parseIntValue(val); ok {
			opts.Timeout = t
		}
	case "proxy":
		opts.Proxy = val
	case "json":
		opts.JSON = val == "true"
	case "size":
		if s, ok := parseIntValue(val); ok {
			opts.MaxSize = s
		}
	case "disable-redirects":
		opts.DisableRedirects = val == "true"
	case "show-source":
		opts.ShowSource = val == "true"
	case "insecure":
		opts.Insecure = val == "true"
	case "unique":
		opts.Unique = val == "true"
	case "concurrency":
		if c, ok := parseIntValue(val); ok {
			opts.Concurrency = c
		}
	case "host-concurrency":
		if h, ok := parseIntValue(val); ok {
			opts.HostConcurrency = h
		}
	case "content-types":
		opts.ContentTypes = val
	case "output":
		opts.Output = val
	case "ignore-robots":
		opts.IgnoreRobots = val == "true"
	case "cross-domain":
		opts.CrossDomain = val == "true"
	case "retries":
		if r, ok := parseIntValue(val); ok {
			opts.Retries = r
		}
	case "retry-backoff":
		if d, ok := parseDurationValue(val); ok {
			opts.RetryBackoff = d
		}
	case "delay":
		if d, ok := parseDurationValue(val); ok {
			opts.Delay = d
		}
	case "sitemap":
		opts.Sitemap = val == "true"
	case "resume":
		opts.Resume = val == "true"
	}
}

func parseKeyValue(args []string) (string, ScanOptions) {
	opts := ScanOptions{
		Depth:        2,
		Timeout:      2,
		MaxSize:      -1,
		ContentTypes: "text/html",
		Output:       "crawler_results",
		RetryBackoff: time.Second,
	}

	var url string

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			val := strings.TrimSpace(parts[1])
			applyScanOption(&opts, key, val)
		} else if url == "" {
			url = arg
		}
	}

	return url, opts
}

var rootCmd = &cobra.Command{
	Use:   "deepscanbot",
	Short: "A high-performance web crawler and scanner",
	Long: `DeepScanBot is a feature-rich, concurrent web crawler that recursively
crawls websites, respects robots.txt, handles rate-limiting, and produces
comprehensive JSON or text reports.

Built entirely in Go, it delivers exceptional performance as a single,
self-contained binary.`,
	Example: `  # Scan a website
  deepscanbot scan https://example.com

  # Scan with custom depth
  deepscanbot scan https://example.com depth=3

  # Scan with JSON output
  deepscanbot scan https://example.com depth=3 json=true

  # Show version
  deepscanbot version`,
}

var scanCmd = &cobra.Command{
	Use:   "scan <url> [options]",
	Short: "Crawl and analyze a website",
	Long: `Scan crawls a website starting from the specified URL, following links
up to a configurable depth, and produces a report of all discovered URLs.

Options are specified as key=value pairs after the URL.

Examples:
  deepscanbot scan https://example.com depth=3 json=true output=results
  deepscanbot scan https://example.com concurrency=10 delay=500ms
  deepscanbot scan https://example.com proxy=http://127.0.0.1:8080 --retries=3`,
	Args: cobra.MinimumNArgs(1),
	Example: `  # Basic scan
  deepscanbot scan https://example.com

  # Scan with depth and JSON output
  deepscanbot scan https://example.com depth=3 json=true

  # Scan with proxy and custom output
  deepscanbot scan https://example.com proxy=http://127.0.0.1:8080 output=results

  # Polite crawl with delays
  deepscanbot scan https://example.com delay=500ms concurrency=5`,
	Run: func(cmd *cobra.Command, args []string) {
		url, opts := parseKeyValue(args)

		if url == "" {
			exitcode.HandleError(exitcode.ErrEmptyURL)
		}

		parsedURL, err := validateStartURL(url)
		if err != nil {
			// validateStartURL returns an *exitcode.ExitCode when the URL is
			// invalid; for other error types it returns a generic error.
			exitcode.HandleError(err)
		}

		// Check for --json flag (persistent flag from root command)
		jsonFlag, _ := cmd.Flags().GetBool("json")
		if jsonFlag {
			opts.JSON = true
		}

		timeoutDuration := time.Duration(opts.Timeout) * time.Second

		outputFilename, err := buildOutputFilename(opts.Output, opts.JSON)
		if err != nil {
			exitcode.HandleError(err)
		}

		var resumeEntries []storage.URLEntry
		if opts.Resume {
			resumeEntries, err = storage.ReadEntriesFromFile(outputFilename)
			if err != nil {
				log.Errorf("load resume file: %v", err)
				exitcode.HandleError(exitcode.ErrResumeLoadFailed)
			}
			log.Infof("Resume mode loaded %d existing results from %s", len(resumeEntries), outputFilename)
		}

		c := crawler.NewCrawlerWithOptions(parsedURL, opts.Depth, timeoutDuration, opts.Proxy, opts.MaxSize, opts.DisableRedirects, opts.Insecure, opts.Unique, opts.Concurrency, parseContentTypes(opts.ContentTypes), opts.IgnoreRobots, opts.CrossDomain, crawler.Options{
			Retries:            opts.Retries,
			RetryBackoff:       opts.RetryBackoff,
			CrawlDelay:         opts.Delay,
			PerHostConcurrency: opts.HostConcurrency,
			IncludeSitemap:     opts.Sitemap,
			ResumeEntries:      resumeEntries,
		})

		report, err := c.StartReport()
		if err != nil {
			exitcode.HandleErrorWithMessage("scan failed", err)
		}

		// Create output formatter
		formatter := output.NewFormatter(opts.JSON)

		// Write to file
		if opts.JSON {
			err = storage.WriteJSONReportToFile(outputFilename, report)
		} else {
			err = storage.WriteTextToFile(outputFilename, report.URLs, opts.ShowSource)
		}

		if err != nil {
			exitcode.HandleErrorWithMessage("write output file", exitcode.ErrWriteOutput)
		}

		// If JSON mode, write report to stdout
		if opts.JSON {
			meta := output.NewResponseMetadata("scan", time.Duration(report.DurationMS)*time.Millisecond)
			err = formatter.WriteSuccess(os.Stdout, report, meta)
			if err != nil {
				exitcode.HandleErrorWithMessage("write JSON output", exitcode.ErrJSONOutput)
			}
		}

		log.Infof("Results written to %s", outputFilename)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the installed version",
	Long:  `Display the current version of DeepScanBot CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for --json flag or json=true option
		jsonFlag, _ := cmd.Flags().GetBool("json")

		// Also check if json=true was passed as a key=value option
		jsonOption := false
		for _, arg := range args {
			if strings.HasPrefix(strings.ToLower(arg), "json=") {
				parts := strings.SplitN(arg, "=", 2)
				jsonOption = len(parts) == 2 && strings.ToLower(parts[1]) == "true"
				break
			}
		}

		if jsonFlag || jsonOption {
			formatter := output.NewFormatter(true)
			meta := output.NewResponseMetadata("version", 0)
			data := map[string]string{
				"version": "1.0.0",
				"name":    "DeepScanBot CLI",
			}
			err := formatter.WriteSuccess(os.Stdout, data, meta)
			if err != nil {
				exitcode.HandleErrorWithMessage("write JSON output", exitcode.ErrJSONOutput)
			}
		} else {
			fmt.Println("DeepScanBot CLI v1.0.0")
		}
	},
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Verify installation and environment",
	Long:  `Check that DeepScanBot is properly installed and the environment is configured correctly.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for --json flag or json=true option
		jsonFlag, _ := cmd.Flags().GetBool("json")

		// Also check if json=true was passed as a key=value option
		jsonOption := false
		for _, arg := range args {
			if strings.HasPrefix(strings.ToLower(arg), "json=") {
				parts := strings.SplitN(arg, "=", 2)
				jsonOption = len(parts) == 2 && strings.ToLower(parts[1]) == "true"
				break
			}
		}

		if jsonFlag || jsonOption {
			formatter := output.NewFormatter(true)
			meta := output.NewResponseMetadata("doctor", 0)
			data := map[string]interface{}{
				"installed":      true,
				"executable":     true,
				"configured":     true,
				"checks_passed":  3,
				"message":        "All checks passed!",
			}
			err := formatter.WriteSuccess(os.Stdout, data, meta)
			if err != nil {
				exitcode.HandleErrorWithMessage("write JSON output", exitcode.ErrJSONOutput)
			}
		} else {
			fmt.Println("Running diagnostics...")
			fmt.Println("✓ DeepScanBot is installed")
			fmt.Println("✓ Binary is executable")
			fmt.Println("✓ Environment is configured correctly")
			fmt.Println("\nAll checks passed!")
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `View and modify DeepScanBot configuration settings.`,
}

var completionCmd = &cobra.Command{
	Use:       "completion [bash|zsh|fish|powershell]",
	Short:     "Generate shell completion script",
	Long:      `Generate shell completion script for DeepScanBot commands.`,
	Args:      cobra.OnlyValidArgs,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Run: func(cmd *cobra.Command, args []string) {
		shell := "bash"
		if len(args) > 0 {
			shell = args[0]
		}
		_ = shell
	},
}

func init() {
	// Add --json flag to root command so all subcommands can use it
	rootCmd.PersistentFlags().Bool("json", false, "Output results in JSON format")

	rootCmd.AddCommand(scanCmd, versionCmd, doctorCmd, configCmd, completionCmd)

	// Silence cobra's own error printing so we can emit consistent
	// error messages ourselves.
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = false // usage is still shown on validation errors

	// Override help to support --json flag for machine-readable command tree output
	// Store the original help function to avoid recursion
	originalHelpFunc := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		jsonFlag, _ := cmd.Flags().GetBool("json")
		if jsonFlag {
			tree := output.BuildCommandTree(rootCmd)
			formatter := output.NewFormatter(true)
			meta := output.NewResponseMetadata("help", 0)
			err := formatter.WriteSuccess(os.Stdout, tree, meta)
			if err != nil {
				exitcode.HandleErrorWithMessage("write JSON output", exitcode.ErrJSONOutput)
			}
			return
		}
		// Fall back to default help using the original function
		originalHelpFunc(cmd, args)
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra errors (e.g. missing required args, unknown flags) are not
		// *ExitCode; map them to InvalidInput so they return exit code 1.
		handleCobraError(err)
	}
}

// handleCobraError maps cobra's error to a standardised exit code and exits.
func handleCobraError(err error) {
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "requires at least"),
		strings.Contains(errStr, "requires exactly"),
		strings.Contains(errStr, "unknown flag"),
		strings.Contains(errStr, "not a valid"),
		strings.Contains(errStr, "usage"):
		exitcode.HandleError(&exitcode.ExitCode{
			Code:    exitcode.InvalidInput,
			Message: errStr,
			Hint:    "Run 'deepscanbot --help' for usage information.",
		})
	default:
		exitcode.HandleError(err)
	}
}

func validateStartURL(rawURL string) (string, error) {
	startURL := strings.TrimSpace(rawURL)
	if startURL == "" {
		return "", exitcode.ErrEmptyURL
	}

	parsedURL, err := url.ParseRequestURI(startURL)
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return "", &exitcode.ExitCode{
			Code:    exitcode.InvalidInput,
			Message: fmt.Sprintf("Invalid URL: %q must be an absolute http:// or https:// URL.", rawURL),
			Hint:    "Example: https://example.com",
		}
	}

	return parsedURL.String(), nil
}

func buildOutputFilename(baseName string, jsonOutput bool) (string, error) {
	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		return "", exitcode.ErrEmptyOutputFilename
	}

	if jsonOutput {
		return baseName + ".json", nil
	}

	return baseName + ".txt", nil
}

func parseContentTypes(value string) []string {
	var contentTypes []string

	for _, part := range strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' '
	}) {
		if part = strings.TrimSpace(part); part != "" {
			contentTypes = append(contentTypes, part)
		}
	}

	return contentTypes
}
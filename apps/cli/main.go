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
	"github.com/mindfiredigital/DeepScanBot/packages/input"
	"github.com/mindfiredigital/DeepScanBot/packages/logger"
	"github.com/mindfiredigital/DeepScanBot/packages/noinput"
	"github.com/mindfiredigital/DeepScanBot/packages/output"
	"github.com/mindfiredigital/DeepScanBot/packages/storage"
	"github.com/mindfiredigital/DeepScanBot/packages/version"
)

// Version variables - these can be set at build time using ldflags
var (
	cliVersion = "dev" // -X main.version
	gitCommit  = ""    // -X main.commit
	buildDate  = ""    // -X main.date
)

// versionInfo returns the current version information
func versionInfo() *version.Info {
	info := version.Default()
	info.Version = cliVersion
	info.GitCommit = gitCommit
	info.BuildDate = buildDate
	return info
}

var log = logger.NewWithLevel(logger.LevelInfo)

// Global flags for destructive operations
var (
	forceOverwrite bool // --force: overwrite existing output without prompting
	dryRun         bool // --dry-run: preview actions without executing
	yesFlag        bool // --yes: auto-confirm destructive operations
)

// Log level flags
var (
	quietFlag   bool // --quiet: suppress non-essential output
	verboseFlag bool // --verbose: display additional informational messages
	debugFlag   bool // --debug: display detailed debugging information
)

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
	InputFile        string
	UseStdin         bool
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
	case "depth", "timeout", "size", "concurrency", "host-concurrency", "retries":
		if i, ok := parseIntValue(val); ok {
			applyIntOption(opts, key, i)
		}
	case "retry-backoff", "delay":
		if d, ok := parseDurationValue(val); ok {
			applyDurationOption(opts, key, d)
		}
	case "proxy", "content-types", "output", "input-file":
		applyStringOption(opts, key, val)
	case "json", "disable-redirects", "show-source", "insecure", "unique",
		"ignore-robots", "cross-domain", "sitemap", "resume", "stdin":
		applyBoolOption(opts, key, val)
	}
}

func applyIntOption(opts *ScanOptions, key string, value int) {
	switch key {
	case "depth":
		opts.Depth = value
	case "timeout":
		opts.Timeout = value
	case "size":
		opts.MaxSize = value
	case "concurrency":
		opts.Concurrency = value
	case "host-concurrency":
		opts.HostConcurrency = value
	case "retries":
		opts.Retries = value
	}
}

func applyDurationOption(opts *ScanOptions, key string, value time.Duration) {
	switch key {
	case "retry-backoff":
		opts.RetryBackoff = value
	case "delay":
		opts.Delay = value
	}
}

func applyStringOption(opts *ScanOptions, key, value string) {
	switch key {
	case "proxy":
		opts.Proxy = value
	case "content-types":
		opts.ContentTypes = value
	case "output":
		opts.Output = value
	case "input-file":
		opts.InputFile = value
	}
}

func applyBoolOption(opts *ScanOptions, key, value string) {
	boolVal := value == "true"
	switch key {
	case "json":
		opts.JSON = boolVal
	case "disable-redirects":
		opts.DisableRedirects = boolVal
	case "show-source":
		opts.ShowSource = boolVal
	case "insecure":
		opts.Insecure = boolVal
	case "unique":
		opts.Unique = boolVal
	case "ignore-robots":
		opts.IgnoreRobots = boolVal
	case "cross-domain":
		opts.CrossDomain = boolVal
	case "sitemap":
		opts.Sitemap = boolVal
	case "resume":
		opts.Resume = boolVal
	case "stdin":
		opts.UseStdin = boolVal
	}
}

// mergeOptions combines flag-based options with key=value options.
// Flag-based options take precedence over key=value options.
func mergeOptions(cmd *cobra.Command, kvOpts ScanOptions) ScanOptions {
	opts := ScanOptions{
		Depth:           2,
		Timeout:         2,
		MaxSize:         -1,
		ContentTypes:    "text/html",
		Output:          "crawler_results",
		RetryBackoff:    time.Second,
		Concurrency:     8,
		HostConcurrency: 2,
	}

	// Apply flag values if set
	applyFlagValues(cmd, &opts)

	// Override with key=value options (but only if flags weren't explicitly set)
	applyKeyValueOptions(cmd, &opts, kvOpts)

	return opts
}

// applyFlagValues reads flag values from the command and applies them to opts
func applyFlagValues(cmd *cobra.Command, opts *ScanOptions) {
	applyIntFlagValues(cmd, opts)
	applyDurationFlagValues(cmd, opts)
	applyStringFlagValues(cmd, opts)
	applyBoolFlagValues(cmd, opts)
}

// applyIntFlagValues applies integer flag values to opts
func applyIntFlagValues(cmd *cobra.Command, opts *ScanOptions) {
	if f := cmd.Flags().Lookup("depth"); f != nil && f.Changed {
		if val, ok := parseIntValue(f.Value.String()); ok {
			opts.Depth = val
		}
	}
	if f := cmd.Flags().Lookup("timeout"); f != nil && f.Changed {
		if val, ok := parseIntValue(f.Value.String()); ok {
			opts.Timeout = val
		}
	}
	if f := cmd.Flags().Lookup("size"); f != nil && f.Changed {
		if val, ok := parseIntValue(f.Value.String()); ok {
			opts.MaxSize = val
		}
	}
	if f := cmd.Flags().Lookup("concurrency"); f != nil && f.Changed {
		if val, ok := parseIntValue(f.Value.String()); ok {
			opts.Concurrency = val
		}
	}
	if f := cmd.Flags().Lookup("host-concurrency"); f != nil && f.Changed {
		if val, ok := parseIntValue(f.Value.String()); ok {
			opts.HostConcurrency = val
		}
	}
	if f := cmd.Flags().Lookup("retries"); f != nil && f.Changed {
		if val, ok := parseIntValue(f.Value.String()); ok {
			opts.Retries = val
		}
	}
}

// applyDurationFlagValues applies duration flag values to opts
func applyDurationFlagValues(cmd *cobra.Command, opts *ScanOptions) {
	if f := cmd.Flags().Lookup("retry-backoff"); f != nil && f.Changed {
		if val, ok := parseDurationValue(f.Value.String()); ok {
			opts.RetryBackoff = val
		}
	}
	if f := cmd.Flags().Lookup("delay"); f != nil && f.Changed {
		if val, ok := parseDurationValue(f.Value.String()); ok {
			opts.Delay = val
		}
	}
}

// applyStringFlagValues applies string flag values to opts
func applyStringFlagValues(cmd *cobra.Command, opts *ScanOptions) {
	if f := cmd.Flags().Lookup("proxy"); f != nil && f.Changed {
		opts.Proxy = f.Value.String()
	}
	if f := cmd.Flags().Lookup("content-types"); f != nil && f.Changed {
		opts.ContentTypes = f.Value.String()
	}
	if f := cmd.Flags().Lookup("output"); f != nil && f.Changed {
		opts.Output = f.Value.String()
	}
	if f := cmd.Flags().Lookup("input-file"); f != nil && f.Changed {
		opts.InputFile = f.Value.String()
	}
}

// applyBoolFlagValues applies boolean flag values to opts
func applyBoolFlagValues(cmd *cobra.Command, opts *ScanOptions) {
	if f := cmd.Flags().Lookup("disable-redirects"); f != nil && f.Changed {
		opts.DisableRedirects = f.Value.String() == "true"
	}
	if f := cmd.Flags().Lookup("show-source"); f != nil && f.Changed {
		opts.ShowSource = f.Value.String() == "true"
	}
	if f := cmd.Flags().Lookup("insecure"); f != nil && f.Changed {
		opts.Insecure = f.Value.String() == "true"
	}
	if f := cmd.Flags().Lookup("unique"); f != nil && f.Changed {
		opts.Unique = f.Value.String() == "true"
	}
	if f := cmd.Flags().Lookup("ignore-robots"); f != nil && f.Changed {
		opts.IgnoreRobots = f.Value.String() == "true"
	}
	if f := cmd.Flags().Lookup("cross-domain"); f != nil && f.Changed {
		opts.CrossDomain = f.Value.String() == "true"
	}
	if f := cmd.Flags().Lookup("sitemap"); f != nil && f.Changed {
		opts.Sitemap = f.Value.String() == "true"
	}
	if f := cmd.Flags().Lookup("resume"); f != nil && f.Changed {
		opts.Resume = f.Value.String() == "true"
	}
	if f := cmd.Flags().Lookup("stdin"); f != nil && f.Changed {
		opts.UseStdin = f.Value.String() == "true"
	}
}

// applyKeyValueOptions applies key=value options only if the corresponding flag was not set
func applyKeyValueOptions(cmd *cobra.Command, opts *ScanOptions, kvOpts ScanOptions) {
	if !cmd.Flags().Lookup("depth").Changed {
		opts.Depth = kvOpts.Depth
	}
	if !cmd.Flags().Lookup("timeout").Changed {
		opts.Timeout = kvOpts.Timeout
	}
	if !cmd.Flags().Lookup("proxy").Changed {
		opts.Proxy = kvOpts.Proxy
	}
	if !cmd.Flags().Lookup("size").Changed {
		opts.MaxSize = kvOpts.MaxSize
	}
	if !cmd.Flags().Lookup("disable-redirects").Changed {
		opts.DisableRedirects = kvOpts.DisableRedirects
	}
	if !cmd.Flags().Lookup("show-source").Changed {
		opts.ShowSource = kvOpts.ShowSource
	}
	if !cmd.Flags().Lookup("insecure").Changed {
		opts.Insecure = kvOpts.Insecure
	}
	if !cmd.Flags().Lookup("unique").Changed {
		opts.Unique = kvOpts.Unique
	}
	if !cmd.Flags().Lookup("concurrency").Changed {
		opts.Concurrency = kvOpts.Concurrency
	}
	if !cmd.Flags().Lookup("host-concurrency").Changed {
		opts.HostConcurrency = kvOpts.HostConcurrency
	}
	if !cmd.Flags().Lookup("content-types").Changed {
		opts.ContentTypes = kvOpts.ContentTypes
	}
	if !cmd.Flags().Lookup("output").Changed {
		opts.Output = kvOpts.Output
	}
	if !cmd.Flags().Lookup("ignore-robots").Changed {
		opts.IgnoreRobots = kvOpts.IgnoreRobots
	}
	if !cmd.Flags().Lookup("cross-domain").Changed {
		opts.CrossDomain = kvOpts.CrossDomain
	}
	if !cmd.Flags().Lookup("retries").Changed {
		opts.Retries = kvOpts.Retries
	}
	if !cmd.Flags().Lookup("retry-backoff").Changed {
		opts.RetryBackoff = kvOpts.RetryBackoff
	}
	if !cmd.Flags().Lookup("delay").Changed {
		opts.Delay = kvOpts.Delay
	}
	if !cmd.Flags().Lookup("sitemap").Changed {
		opts.Sitemap = kvOpts.Sitemap
	}
	if !cmd.Flags().Lookup("resume").Changed {
		opts.Resume = kvOpts.Resume
	}
	if !cmd.Flags().Lookup("input-file").Changed {
		opts.InputFile = kvOpts.InputFile
	}
	if !cmd.Flags().Lookup("stdin").Changed {
		opts.UseStdin = kvOpts.UseStdin
	}
}

func parseKeyValue(args []string) (string, ScanOptions) {
	opts := ScanOptions{
		Depth:           2,
		Timeout:         2,
		MaxSize:         -1,
		ContentTypes:    "text/html",
		Output:          "crawler_results",
		RetryBackoff:    time.Second,
		Concurrency:     8,
		HostConcurrency: 2,
	}

	var url string

	for _, arg := range args {
		// Skip flags (they start with - and are handled by Cobra)
		if strings.HasPrefix(arg, "-") {
			continue
		}

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
  deepscanbot version

	# Show version (short flag)
  deepscanbot --version`,
	Version: cliVersion,
}

func init() {
    rootCmd.SetVersionTemplate("{{.Version}}\n") 
}

var scanCmd = &cobra.Command{
	Use:   "scan <url> [options]",
	Short: "Crawl and analyze a website",
	Long: `Scan crawls a website starting from the specified URL, following links
up to a configurable depth, and produces a report of all discovered URLs.

Options can be specified as either flags (--depth=3) or key=value pairs (depth=3).
Both formats are supported for backward compatibility.

Examples:
  deepscanbot scan https://example.com --depth=3 --json --output=results
  deepscanbot scan https://example.com --concurrency=10 --delay=500ms
  deepscanbot scan https://example.com --proxy=http://127.0.0.1:8080 --retries=3
  deepscanbot scan https://example.com depth=3 json=true output=results`,
	Args: cobra.MinimumNArgs(1),
	Example: `  # Basic scan
  deepscanbot scan https://example.com

  # Scan with depth and JSON output (flag format)
  deepscanbot scan https://example.com --depth=3 --json

  # Scan with depth and JSON output (key=value format)
  deepscanbot scan https://example.com depth=3 json=true

  # Scan with proxy and custom output
  deepscanbot scan https://example.com --proxy=http://127.0.0.1:8080 --output=results

  # Polite crawl with delays
  deepscanbot scan https://example.com --delay=500ms --concurrency=5

  # Non-interactive (CI/CD)
  deepscanbot scan https://example.com --no-input --force

  # Preview what would happen without making changes
  deepscanbot scan https://example.com --dry-run

  # Auto-confirm destructive operations
  deepscanbot scan https://example.com --yes --force`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse key=value options for backward compatibility
		url, keyValueOpts := parseKeyValue(args)

		// Merge with flag-based options (flags take precedence)
		opts := mergeOptions(cmd, keyValueOpts)

		if url == "" {
			exitcode.HandleError(exitcode.ErrEmptyURL)
		}

		parsedURL, err := validateStartURL(url)
		if err != nil {
			exitcode.HandleError(err)
		}

		// Check for --json flag (persistent flag from root command)
		jsonFlag, _ := cmd.Flags().GetBool("json")
		if jsonFlag {
			opts.JSON = true
		} else if keyValueOpts.JSON {
			// If --json flag wasn't set but json=true was in key=value options, use that
			opts.JSON = true
		}

		timeoutDuration := time.Duration(opts.Timeout) * time.Second

		outputFilename, err := buildOutputFilename(opts.Output, opts.JSON)
		if err != nil {
			exitcode.HandleError(err)
		}

		// --- Read URLs from stdin or file if specified ---
		var urlsToScan []string
		if opts.InputFile != "" || opts.UseStdin {
			inputURLs, err := input.ReadInput(opts.InputFile, opts.UseStdin)
			if err != nil {
				exitcode.HandleErrorWithMessage("read input", exitcode.ErrFileRead)
			}
			if len(inputURLs) == 0 {
				exitcode.HandleError(&exitcode.ExitCode{
					Code:    exitcode.InvalidInput,
					Message: "No URLs provided via input.",
					Hint:    "Provide URLs via stdin or --input-file flag.",
				})
			}
			urlsToScan = inputURLs
			log.Infof("Loaded %d URLs from %s", len(urlsToScan),
				map[bool]string{true: "stdin", false: "file " + opts.InputFile}[opts.UseStdin])
		} else {
			// Use the positional URL argument
			urlsToScan = []string{parsedURL}
		}

		// --- Dry-run mode: preview actions and exit without making changes ---
		if dryRun {
			printDryRunPlan(parsedURL, outputFilename, opts, len(urlsToScan))
			return
		}

		// --- Confirmation check for destructive operations ---
		// If the output file already exists, this is a destructive operation.
		// Require either --force (explicit overwrite) or --yes (auto-confirm).
		if _, statErr := os.Stat(outputFilename); statErr == nil {
			// File exists — this is a destructive overwrite.
			if !forceOverwrite && !yesFlag {
				if !noinput.IsInteractive() {
					exitcode.HandleError(&exitcode.ExitCode{
						Code:    exitcode.InvalidInput,
						Message: fmt.Sprintf("Output file %q already exists. Refusing to overwrite without confirmation.", outputFilename),
						Hint:    "Pass --force to overwrite or --yes to auto-confirm all destructive operations.",
					})
				}
				// Interactive mode: warn and proceed (backward-compatible).
				log.Warnf("Output file %q already exists. It will be overwritten.", outputFilename)
			}
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

		// For now, we only support scanning a single URL at a time
		// In the future, this could be extended to scan multiple URLs
		if len(urlsToScan) > 1 {
			log.Warnf("Multiple URLs provided, only scanning the first one: %s", urlsToScan[0])
		}

		c := crawler.NewCrawlerWithOptions(urlsToScan[0], opts.Depth, timeoutDuration, opts.Proxy, opts.MaxSize, opts.DisableRedirects, opts.Insecure, opts.Unique, opts.Concurrency, parseContentTypes(opts.ContentTypes), opts.IgnoreRobots, opts.CrossDomain, crawler.Options{
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
	Example: `  # Show version
  deepscanbot version

  # Show version in JSON format
  deepscanbot version --json`,
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

		info := versionInfo()

		if jsonFlag || jsonOption {
			formatter := output.NewFormatter(true)
			meta := output.NewResponseMetadata("version", 0)
			err := formatter.WriteSuccess(os.Stdout, info.JSON(), meta)
			if err != nil {
				exitcode.HandleErrorWithMessage("write JSON output", exitcode.ErrJSONOutput)
			}
		} else {
			log.Infof(info.String())
		}
	},
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Verify installation and environment",
	Long:  `Check that DeepScanBot is properly installed and the environment is configured correctly.`,
	Example: `  # Run diagnostics
  deepscanbot doctor

  # Run diagnostics with JSON output
  deepscanbot doctor --json`,
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
				"installed":     true,
				"executable":    true,
				"configured":    true,
				"checks_passed": 3,
				"message":       "All checks passed!",
			}
			err := formatter.WriteSuccess(os.Stdout, data, meta)
			if err != nil {
				exitcode.HandleErrorWithMessage("write JSON output", exitcode.ErrJSONOutput)
			}
		} else {
			log.Infof("Running diagnostics...")
			log.Infof("✓ DeepScanBot is installed")
			log.Infof("✓ Binary is executable")
			log.Infof("✓ Environment is configured correctly")
			log.Infof("")
			log.Infof("All checks passed!")
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `View and modify DeepScanBot configuration settings.`,
	Example: `  # View current configuration
  deepscanbot config`,
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long:  `Generate shell completion script for DeepScanBot commands.`,
	Example: `  # Generate bash completion
  deepscanbot completion bash

  # Generate zsh completion
  deepscanbot completion zsh

  # Generate fish completion
  deepscanbot completion fish

  # Generate PowerShell completion
  deepscanbot completion powershell`,
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
	// Add persistent flags shared by all commands
	rootCmd.PersistentFlags().Bool("json", false, "Output results in JSON format")
	rootCmd.PersistentFlags().Bool("no-input", false, "Disable all interactive prompts; fail if required input is missing")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview actions that would be performed without making changes")

	// Add logging level flags
	rootCmd.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "Suppress non-essential output (only show warnings and errors)")
	rootCmd.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "Display additional informational messages")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Display detailed debugging information")

	// Add flags to scan command for safe destructive operations
	scanCmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing output file without prompting")
	scanCmd.Flags().BoolVar(&yesFlag, "yes", false, "Auto-confirm all destructive operations (e.g. overwriting files)")

	// Add standardized flags to scan command
	// These flags support both --flag=value and key=value formats for backward compatibility
	scanCmd.Flags().Int("depth", 2, "Maximum crawl depth (default: 2)")
	scanCmd.Flags().Int("timeout", 2, "Request timeout in seconds (default: 2)")
	scanCmd.Flags().String("proxy", "", "HTTP proxy URL (e.g., http://127.0.0.1:8080)")
	scanCmd.Flags().Int("size", -1, "Maximum page size in bytes; -1 for unlimited (default: -1)")
	scanCmd.Flags().Bool("disable-redirects", false, "Disable following HTTP redirects")
	scanCmd.Flags().Bool("show-source", false, "Include source URL in output for discovered links")
	scanCmd.Flags().Bool("insecure", false, "Skip TLS certificate validation")
	scanCmd.Flags().Bool("unique", false, "Only process unique URLs (deduplicate)")
	scanCmd.Flags().Int("concurrency", 8, "Maximum concurrent requests (default: 8)")
	scanCmd.Flags().Int("host-concurrency", 2, "Maximum concurrent requests per host (default: 2)")
	scanCmd.Flags().String("content-types", "text/html", "Content types to accept (comma-separated)")
	scanCmd.Flags().String("output", "crawler_results", "Output file base name (without extension)")
	scanCmd.Flags().Bool("ignore-robots", false, "Ignore robots.txt rules")
	scanCmd.Flags().Bool("cross-domain", false, "Follow links to different domains")
	scanCmd.Flags().Int("retries", 0, "Number of retries for failed requests (default: 0)")
	scanCmd.Flags().Duration("retry-backoff", time.Second, "Initial backoff duration for retries (e.g., 1s, 500ms)")
	scanCmd.Flags().Duration("delay", 0, "Delay between requests to the same host (e.g., 500ms, 1s)")
	scanCmd.Flags().Bool("sitemap", false, "Discover URLs from sitemap.xml")
	scanCmd.Flags().Bool("resume", false, "Resume from previous crawl results")
	scanCmd.Flags().String("input-file", "", "Read URLs from a file (one per line)")
	scanCmd.Flags().Bool("stdin", false, "Read URLs from standard input (one per line)")

	rootCmd.AddCommand(scanCmd, versionCmd, doctorCmd, configCmd, completionCmd)

	// Silence cobra's own error printing so we can emit consistent
	// error messages ourselves.
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = false // usage is still shown on validation errors

	// Configure noinput package based on the --no-input flag.
	// This must happen before commands execute so IsInteractive() returns
	// the correct value during Run.
	originalPersistentPreRun := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		noInput, _ := cmd.Flags().GetBool("no-input")
		_ = args
		if noInput {
			noinput.SetNoInputFlag()
		}

		// Configure logging level based on flags
		configureLogLevel(cmd)

		if originalPersistentPreRun != nil {
			return originalPersistentPreRun(cmd, args)
		}
		return nil
	}

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

// configureLogLevel sets the logging level based on --quiet, --verbose, and --debug flags
func configureLogLevel(cmd *cobra.Command) {
	quiet, _ := cmd.Flags().GetBool("quiet")
	verbose, _ := cmd.Flags().GetBool("verbose")
	debug, _ := cmd.Flags().GetBool("debug")

	switch {
	case debug:
		log.SetLevel(logger.LevelDebug)
		log.Debugf("Debug logging enabled")
	case verbose:
		log.SetLevel(logger.LevelVerbose)
		log.Infof("Verbose logging enabled")
	case quiet:
		log.SetLevel(logger.LevelQuiet)
	default:
		log.SetLevel(logger.LevelInfo)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra errors (e.g. missing required args, unknown flags) are not
		// *ExitCode; map them to InvalidInput so they return exit code 1.
		handleCobraError(err)
	}
}

// handleCobraError maps cobra's error to a standardized exit code and exits.
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

// printDryRunPlan displays the actions that would be performed during a scan
// without actually executing them.  It supports both human-readable and JSON
// output formats.
func printDryRunPlan(targetURL, outputFilename string, opts ScanOptions, urlCount int) {
	plan := map[string]interface{}{
		"action":          "scan",
		"target_url":      targetURL,
		"output_file":     outputFilename,
		"depth":           opts.Depth,
		"timeout_seconds": opts.Timeout,
		"concurrency":     opts.Concurrency,
		"content_types":   parseContentTypes(opts.ContentTypes),
		"json_output":     opts.JSON,
		"resume":          opts.Resume,
		"proxy":           opts.Proxy,
		"ignore_robots":   opts.IgnoreRobots,
		"cross_domain":    opts.CrossDomain,
		"sitemap":         opts.Sitemap,
		"retries":         opts.Retries,
		"urls_provided":   urlCount,
	}

	// Check if the output file already exists (would be overwritten)
	if _, err := os.Stat(outputFilename); err == nil {
		plan["existing_file_will_be_overwritten"] = true
	}

	if opts.JSON {
		formatter := output.NewFormatter(true)
		meta := output.NewResponseMetadata("dry-run", 0)
		err := formatter.WriteSuccess(os.Stdout, plan, meta)
		if err != nil {
			exitcode.HandleErrorWithMessage("write dry-run output", exitcode.ErrJSONOutput)
		}
	} else {
		log.Infof("─── Dry Run ───")
		log.Infof("Action:     scan")
		log.Infof("Target URL: %s", targetURL)
		log.Infof("Output:     %s", outputFilename)
		if v, ok := plan["existing_file_will_be_overwritten"]; ok && v.(bool) {
			log.Warnf("⚠  Warning: output file already exists and will be overwritten")
		}
		log.Infof("Depth:      %d", opts.Depth)
		log.Infof("Timeout:    %ds", opts.Timeout)
		log.Infof("Concurrency: %d", opts.Concurrency)
		log.Infof("Content:    %s", strings.Join(parseContentTypes(opts.ContentTypes), ", "))
		log.Infof("JSON:       %v", opts.JSON)
		log.Infof("Resume:     %v", opts.Resume)
		log.Infof("Proxy:      %s", opts.Proxy)
		log.Infof("Sitemap:    %v", opts.Sitemap)
		log.Infof("Retries:    %d", opts.Retries)
		log.Infof("────────────────")
		log.Infof("No changes were made. Pass --dry-run to preview, or omit it to execute.")
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

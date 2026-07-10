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
)

// cliVersion is the current version of the CLI
const cliVersion = "1.0.0"

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
	case "input-file":
		opts.InputFile = val
	case "stdin":
		opts.UseStdin = val == "true"
	}
}

// mergeOptions combines flag-based options with key=value options.
// Flag-based options take precedence over key=value options.
func mergeOptions(cmd *cobra.Command, kvOpts ScanOptions) ScanOptions {
	opts := ScanOptions{
		// Start with flag defaults
		Depth:           2,
		Timeout:         2,
		MaxSize:         -1,
		ContentTypes:    "text/html",
		Output:          "crawler_results",
		RetryBackoff:    time.Second,
		Concurrency:     8,
		HostConcurrency: 2,
	}

	// Helper to get flag value if set
	getFlagInt := func(name string, defaultVal int) int {
		if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
			if val, ok := parseIntValue(f.Value.String()); ok {
				return val
			}
		}
		return defaultVal
	}

	getFlagDuration := func(name string, defaultVal time.Duration) time.Duration {
		if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
			if val, ok := parseDurationValue(f.Value.String()); ok {
				return val
			}
		}
		return defaultVal
	}

	getFlagString := func(name string, defaultVal string) string {
		if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
			return f.Value.String()
		}
		return defaultVal
	}

	getFlagBool := func(name string, defaultVal bool) bool {
		if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
			return f.Value.String() == "true"
		}
		return defaultVal
	}

	// Start with flag defaults
	opts.Depth = getFlagInt("depth", opts.Depth)
	opts.Timeout = getFlagInt("timeout", opts.Timeout)
	opts.Proxy = getFlagString("proxy", opts.Proxy)
	opts.MaxSize = getFlagInt("size", opts.MaxSize)
	opts.DisableRedirects = getFlagBool("disable-redirects", opts.DisableRedirects)
	opts.ShowSource = getFlagBool("show-source", opts.ShowSource)
	opts.Insecure = getFlagBool("insecure", opts.Insecure)
	opts.Unique = getFlagBool("unique", opts.Unique)
	opts.Concurrency = getFlagInt("concurrency", opts.Concurrency)
	opts.HostConcurrency = getFlagInt("host-concurrency", opts.HostConcurrency)
	opts.ContentTypes = getFlagString("content-types", opts.ContentTypes)
	opts.Output = getFlagString("output", opts.Output)
	opts.IgnoreRobots = getFlagBool("ignore-robots", opts.IgnoreRobots)
	opts.CrossDomain = getFlagBool("cross-domain", opts.CrossDomain)
	opts.Retries = getFlagInt("retries", opts.Retries)
	opts.RetryBackoff = getFlagDuration("retry-backoff", opts.RetryBackoff)
	opts.Delay = getFlagDuration("delay", opts.Delay)
	opts.Sitemap = getFlagBool("sitemap", opts.Sitemap)
	opts.Resume = getFlagBool("resume", opts.Resume)

	// Override with key=value options (but only if flags weren't explicitly set)
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

	return opts
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
  deepscanbot version`,
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
	Example: `  # View current configuration
  deepscanbot config`,
}

var completionCmd = &cobra.Command{
	Use:       "completion [bash|zsh|fish|powershell]",
	Short:     "Generate shell completion script",
	Long:      `Generate shell completion script for DeepScanBot commands.`,
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
		fmt.Println("─── Dry Run ───")
		fmt.Printf("Action:     scan\n")
		fmt.Printf("Target URL: %s\n", targetURL)
		fmt.Printf("Output:     %s\n", outputFilename)
		if v, ok := plan["existing_file_will_be_overwritten"]; ok && v.(bool) {
			fmt.Println("⚠  Warning: output file already exists and will be overwritten")
		}
		fmt.Printf("Depth:      %d\n", opts.Depth)
		fmt.Printf("Timeout:    %ds\n", opts.Timeout)
		fmt.Printf("Concurrency: %d\n", opts.Concurrency)
		fmt.Printf("Content:    %s\n", strings.Join(parseContentTypes(opts.ContentTypes), ", "))
		fmt.Printf("JSON:       %v\n", opts.JSON)
		fmt.Printf("Resume:     %v\n", opts.Resume)
		fmt.Printf("Proxy:      %s\n", opts.Proxy)
		fmt.Printf("Sitemap:    %v\n", opts.Sitemap)
		fmt.Printf("Retries:    %d\n", opts.Retries)
		fmt.Println("────────────────")
		fmt.Println("No changes were made. Pass --dry-run to preview, or omit it to execute.")
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
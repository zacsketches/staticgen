package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type BaseData struct {
	Year           int
	BuildTimestamp string
	// Add anything you want available to every page:
	// UserName string
	// Env      string
}

func main() {
	var srcDir, outDir, pagesGlob, buildTimestamp string
	flag.StringVar(&srcDir, "src", "./src", "source directory")
	flag.StringVar(&outDir, "out", "./site", "output directory")
	flag.StringVar(&pagesGlob, "glob", "pages/**/*.template.html", "glob for pages within src directory")

	// Load CST timezone
	cst, err := time.LoadLocation("America/Chicago")
	if err != nil {
		fatal(err)
	}
	defaultTimestamp := time.Now().In(cst).Format("2006-01-02 15:04:05 MST")
	flag.StringVar(&buildTimestamp, "timestamp", defaultTimestamp, "build timestamp")
	flag.Parse()

	log.Printf("Starting static site generation...")
	log.Printf("Source directory: %s", srcDir)
	log.Printf("Output directory: %s", outDir)
	log.Printf("Build timestamp: %s", buildTimestamp)

	// Ensure output dir exists
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}

	// Collect page files
	pagePattern := filepath.Join(srcDir, pagesGlob)
	pageFiles, err := filepath.Glob(pagePattern)
	if err != nil {
		fatal(err)
	}
	if len(pageFiles) == 0 {
		fatal(errors.New("no page templates found: " + pagePattern))
	}

	log.Printf("Found %d page template(s) to render", len(pageFiles))

	// Common includes/layouts
	includesGlobs := []string{
		filepath.Join(srcDir, "_includes", "*.html"),
		filepath.Join(srcDir, "_layouts", "*.html"),
	}

	for _, page := range pageFiles {
		if err := renderOne(srcDir, outDir, includesGlobs, page, buildTimestamp); err != nil {
			fatal(err)
		}
	}

	log.Printf("Static site generation completed successfully")
}

func renderOne(srcDir, outDir string, includesGlobs []string, pageFile string, buildTimestamp string) error {
	// Build the full list of template files for this page
	var all []string
	for _, g := range includesGlobs {
		matches, _ := filepath.Glob(g)
		all = append(all, matches...)
	}
	all = append(all, pageFile)

	// Parse as one set so blocks/partials can see each other
	funcs := template.FuncMap{
		// Add helper funcs as needed
		"nowRFC3339": func() string { return time.Now().Format(time.RFC3339) },
	}
	t, err := template.New("root").Funcs(funcs).ParseFiles(all...)
	if err != nil {
		return err
	}

	// Default layout
	layout := "public"

	// Detect optional {{define "layout_name"}}dashboard{{end}} etc.
	if t.Lookup("layout_name") != nil {
		var b strings.Builder
		if err := t.ExecuteTemplate(&b, "layout_name", nil); err == nil {
			name := strings.TrimSpace(b.String())
			if name != "" {
				layout = name
			}
		}
	}

	// Determine output path: pages/foo.template.html -> site/foo.html
	rel, err := filepath.Rel(filepath.Join(srcDir, "pages"), pageFile)
	if err != nil {
		return err
	}
	outName := strings.TrimSuffix(rel, ".template.html") + ".html"
	outPath := filepath.Join(outDir, outName)

	// Log the rendering operation
	log.Printf("Rendering %s -> %s (layout: %s)", pageFile, outPath, layout)

	// Ensure subdirs exist
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}

	// Create/write the file
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	data := BaseData{
		Year:           time.Now().Year(),
		BuildTimestamp: buildTimestamp,
	}
	// Execute the top-level layout so it pulls in header/footer and the page blocks
	if err := t.ExecuteTemplate(f, layout, data); err != nil {
		return fmt.Errorf("%s: executing layout %q: %w", pageFile, layout, err)
	}

	// Log successful completion
	log.Printf("Successfully wrote %s", outPath)

	// Optional: fmt the HTML, minify, etc.
	return nil
}

func fatal(err error) {
	// Print path-aware errors nicely for CI logs
	var perr *fs.PathError
	if errors.As(err, &perr) {
		_, _ = os.Stderr.WriteString(perr.Path + ": " + perr.Err.Error() + "\n")
	} else {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
	}
	os.Exit(1)
}

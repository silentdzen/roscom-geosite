package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: import-loyalsoldier-geosite INPUT.dat OUTPUT_DIR")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputDir := os.Args[2]

	data, err := os.ReadFile(inputPath)
	if err != nil {
		fail("read upstream database: %v", err)
	}

	var db router.GeoSiteList
	if err := proto.Unmarshal(data, &db); err != nil {
		fail("invalid upstream geosite.dat: %v", err)
	}

	wanted := map[string]bool{
		"OPENAI":    false,
		"ANTHROPIC": false,
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fail("create output directory: %v", err)
	}

	for _, site := range db.Entry {
		name := strings.ToUpper(site.CountryCode)
		if _, ok := wanted[name]; !ok {
			continue
		}

		var lines []string

		for _, d := range site.Domain {
			switch d.Type {
			case router.Domain_RootDomain:
				lines = append(lines, "domain:"+d.Value)
			case router.Domain_Full:
				lines = append(lines, "full:"+d.Value)
			case router.Domain_Plain:
				lines = append(lines, "keyword:"+d.Value)
			case router.Domain_Regex:
				lines = append(lines, "regexp:"+d.Value)
			}
		}

		if len(lines) == 0 {
			fail("upstream category %s is empty", name)
		}

		filename := strings.ToLower(name)
		outPath := filepath.Join(outputDir, filename)

		// Merge with existing entries so manual additions survive.
		existing, err := os.ReadFile(outPath)
		if err != nil && !os.IsNotExist(err) {
			fail("read existing %s: %v", outPath, err)
		}
		if err == nil {
			for _, line := range strings.Split(string(existing), "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					lines = append(lines, line)
				}
			}
		}

		// Deduplicate and sort entries.
		seen := make(map[string]struct{}, len(lines))
		merged := make([]string, 0, len(lines))

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if _, exists := seen[line]; exists {
				continue
			}
			seen[line] = struct{}{}
			merged = append(merged, line)
		}

		sort.Strings(merged)

		content := strings.Join(merged, "\n") + "\n"
		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			fail("write %s: %v", outPath, err)
		}

		wanted[name] = true
		fmt.Printf("%s: %d entries written to %s\n",
			filename, len(merged), outPath)
	}

	for name, found := range wanted {
		if !found {
			fail("category %s is missing from upstream geosite.dat", name)
		}
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}

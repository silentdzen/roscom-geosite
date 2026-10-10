package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

func main() {
	if len(os.Args) != 3 {
		panic("usage: import-loyalsoldier-geosite INPUT.dat OUTPUT_DIR")
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var db router.GeoSiteList
	if err := proto.Unmarshal(data, &db); err != nil {
		panic(fmt.Errorf("invalid upstream geosite.dat: %w", err))
	}

	wanted := map[string]bool{
		"OPENAI":        false,
		"GOOGLE-GEMINI": false,
		"ANTHROPIC":     false,
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
			panic("empty upstream category: " + name)
		}

		if err := os.MkdirAll(os.Args[2], 0755); err != nil {
			panic(err)
		}

		out := filepath.Join(os.Args[2], name)
		content := strings.Join(lines, "\n") + "\n"

		if err := os.WriteFile(out, []byte(content), 0644); err != nil {
			panic(err)
		}

		wanted[name] = true
		fmt.Printf("%s: %d entries\n", name, len(lines))
	}

	for name, found := range wanted {
		if !found {
			panic("category missing from upstream database: " + name)
		}
	}
}

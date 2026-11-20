package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
	"google.golang.org/protobuf/proto"
)

const category = "RU-AVAILABLE-ONLY-INSIDE"

func main() {
	if len(os.Args) != 3 {
		panic("usage: import-runetfreedom-geosite INPUT.dat OUTPUT_DIR")
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var db router.GeoSiteList
	if err := proto.Unmarshal(data, &db); err != nil {
		panic(fmt.Errorf("invalid upstream geosite.dat: %w", err))
	}

	var selected *router.GeoSite
	for _, site := range db.Entry {
		if strings.EqualFold(site.CountryCode, category) {
			selected = site
			break
		}
	}

	if selected == nil {
		panic("category missing from upstream database: " + category)
	}

	var lines []string
	for _, d := range selected.Domain {
		switch d.Type {
		case router.Domain_RootDomain:
			lines = append(lines, "domain:"+d.Value)
		case router.Domain_Full:
			lines = append(lines, "full:"+d.Value)
		case router.Domain_Plain:
			lines = append(lines, "keyword:"+d.Value)
		case router.Domain_Regex:
			lines = append(lines, "regexp:"+d.Value)
		default:
			panic(fmt.Sprintf("unknown domain type for %q", d.Value))
		}
	}

	if len(lines) == 0 {
		panic("empty upstream category: " + category)
	}

	if err := os.MkdirAll(os.Args[2], 0755); err != nil {
		panic(err)
	}

	out := filepath.Join(os.Args[2], "ru-available-only-inside")
	content := strings.Join(lines, "\n") + "\n"

	if err := os.WriteFile(out, []byte(content), 0644); err != nil {
		panic(err)
	}

	fmt.Printf("%s: %d entries written to %s\n", category, len(lines), out)
}

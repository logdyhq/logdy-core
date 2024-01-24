package main

import (
	"embed"
	"io/fs"
)

//go:embed all:assets
var assets embed.FS

func Assets() (fs.FS, error) {
	return fs.Sub(assets, "assets")
}

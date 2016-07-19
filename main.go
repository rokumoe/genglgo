package main

import (
	"flag"
	"path/filepath"
)

func main() {
	var (
		optInput   string
		optOutput  string
		optAPI     string
		optProfile string
		optVersion string
	)
	flag.StringVar(&optInput, "input", "res/gl.xml", "input path of gl.xml")
	flag.StringVar(&optOutput, "output", "gl/gl.go", "output path of gl.go")
	flag.StringVar(&optAPI, "api", "gl", "GL API")
	flag.StringVar(&optProfile, "profile", "core", "GL profile[core|compatibility]")
	flag.StringVar(&optVersion, "version", "3.2", "GL version")
	flag.Parse()
	if !flag.Parsed() || flag.NArg() != 0 {
		panic("error flags")
	}
	if optProfile != "" && optProfile != "core" && optProfile != "compatibility" {
		panic("invalid profile")
	}
	outpath, err := filepath.Abs(optOutput)
	if err != nil {
		panic(err)
	}
	if err := generate(optInput, optAPI, optProfile, optVersion, outpath); err != nil {
		panic(err)
	}
}

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
		optVersion string
	)
	flag.StringVar(&optInput, "input", "res/gl.xml", "input path of gl.xml")
	flag.StringVar(&optOutput, "output", "gl/gl.go", "output path of gl.go")
	flag.StringVar(&optAPI, "api", "gl", "GL API")
	flag.StringVar(&optVersion, "version", "3.2", "OpenGL version")
	flag.Parse()
	if !flag.Parsed() || flag.NArg() != 0 {
		panic("error flags")
	}
	outpath, err := filepath.Abs(optOutput)
	if err != nil {
		panic(err)
	}
	if err := generate(optInput, optAPI, optVersion, outpath); err != nil {
		panic(err)
	}
}

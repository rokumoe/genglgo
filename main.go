package main

import (
	"flag"
	"path/filepath"
)

func main() {
	var (
		argin  string
		argout string
		argapi string
		argver string
	)
	flag.StringVar(&argin, "i", "gl.xml", "input path of gl.xml")
	flag.StringVar(&argout, "o", "gl.go", "output path of gl.go")
	flag.StringVar(&argapi, "a", "", "GL API")
	flag.StringVar(&argver, "v", "2.1", "OpenGL version")
	flag.Parse()
	if !flag.Parsed() || flag.NArg() != 0 {
		panic("miss arg")
	}
	outpath, err := filepath.Abs(argout)
	if err != nil {
		panic(err)
	}
	if err := generate(argin, argapi, argver, outpath); err != nil {
		panic(err)
	}
}

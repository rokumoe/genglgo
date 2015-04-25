package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

var template = []string{
	`package gl

//#cgo linux	CFLAGS: -DGL_PLATFORM_LINUX
//#cgo windows	CFLAGS: -DGL_PLATFORM_WINDOWS
//#cgo linux	LDFLAGS: -lGL
//#cgo windows  LDFLAGS: -lopengl32
import "C"

/*
#if defined(GL_PLATFORM_LINUX)
#include <GL/glx.h>

static void* _glgo_GetProcAddress(const char* name) {
	return glXGetProcAddress(name);
}
#elif defined(GL_PLATFORM_WINDOWS)
#include <Windows.h>

static void* _glgo_GetProcAddress(const char* name) {
	return wglGetProcAddress((LPCSTR)name);
}
#else
#error "Unsupport platform"
#endif
*/
import "C"

/*
#if defined(APIENTRY)
#define GLGO_APIENTRY APIENTRY
#elif defined(_STDCALL_SUPPORTED)
#define GLGO_APIENTRY __stdcall
#else
#define GLGO_APIENTRY
#endif

#ifdef far
#undef far
#endif
#ifdef near
#undef near
#endif

#define GLGO_COMMAND_DECL(return_type, command, ...) \
typedef return_type (GLGO_APIENTRY *_glgo_t_##command)(__VA_ARGS__);\
static _glgo_t_##command _glgo_p_##command = 0;\
return_type command(__VA_ARGS__)

#define GLGO_COMMAND_RET(command, ...) \
{\
	return _glgo_p_##command(__VA_ARGS__); \
}

#define GLGO_COMMAND_0RET(command, ...) \
{\
	_glgo_p_##command(__VA_ARGS__); \
}

#define GLGO_COMMAND_GETPROC(command) \
	(_glgo_p_##command = (_glgo_t_##command)_glgo_GetProcAddress(#command))
*/
import "C"
`,
	`/*`,
	`int gl_init()
{`,
	`	return 0;
}*/
import "C"`,
	``,
	``,
	``,
}

type pair struct {
	k string
	v string
}

type command_info struct {
	params  []pair
	rettype string
}

func gen_c_def_type(is_types map[string]bool, api string, types registry_types) string {
	for _, t := range types.type_ {
		if is_types[t.name] && is_same_api(api, t.api) {
			if t.requires != "" {
				is_types[t.requires] = true
			}
		}
	}
	s := ""
	for _, t := range types.type_ {
		if is_types[t.name] && is_same_api(api, t.api) {
			b := 0
			for i := 0; i <= len(t.text); i++ {
				if i == len(t.text) || t.text[i] == '\n' {
					s += "//" + t.text[b:i] + "\n"
					b = i + 1
				}
			}
		}
	}
	return s + "import \"C\"\n"
}

func gen_c_def_command(command string, info command_info) string {
	var (
		params     string
		paramnames string
	)
	for _, p := range info.params {
		if params != "" {
			params += ", "
		}
		if paramnames != "" {
			paramnames += ", "
		}
		params += p.v + " " + p.k
		paramnames += p.k
	}
	def := "GLGO_COMMAND_DECL(" + info.rettype + ", " + command + ", " + params + ")\n"
	if info.rettype == "void" {
		def += "GLGO_COMMAND_0RET"
	} else {
		def += "GLGO_COMMAND_RET"
	}
	def += "(" + command + ", " + paramnames + ")\n"
	return def
}

func gen_c_init_command(command string) string {
	return `	if (GLGO_COMMAND_GETPROC(` + command + `) == NULL) {
		return __LINE__;
	}`
}

func is_same_api(a string, b string) bool {
	return a == b || (a == "" && b == "gl") || (a == "gl" && b == "")
}

func convert(glxml string, api string, number string, glgo string) error {
	registry, err := load_glxml(glxml)
	if err != nil {
		return err
	}
	max_ver, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return err
	}
	var (
		is_types    = make(map[string]bool)
		is_enums    = make(map[string]bool)
		is_commands = make(map[string]bool)
	)
	for _, feature := range registry.feature {
		ver, err := strconv.ParseFloat(feature.number, 64)
		if err != nil {
			return err
		}
		if is_same_api(api, feature.api) && ver <= max_ver {
			for _, require := range feature.require {
				for _, type_ := range require.type_ {
					is_types[type_.name] = true
				}
				for _, enum := range require.enum {
					is_enums[enum.name] = true
				}
				for _, command := range require.command {
					is_commands[command.name] = true
				}
			}
			for _, remove := range feature.remove {
				for _, enum := range remove.enum {
					delete(is_enums, enum.name)
				}
				for _, command := range remove.command {
					delete(is_commands, command.name)
				}
			}
		}
	}
	var (
		enums_map    = make(map[string]string, len(is_enums))
		commands_map = make(map[string]command_info, len(is_commands))
	)
	for _, enums := range registry.enums {
		for _, e := range enums.enum {
			if is_enums[e.name] {
				enums_map[e.name] = e.value
			}
		}
	}
	for _, c := range registry.commands.command {
		if is_commands[c.proto.name] {
			text := c.proto.text
			if c.proto.ptype != "" {
				is_types[c.proto.ptype] = true
			}
			param_list := make([]pair, len(c.param))
			for i, p := range c.param {
				text := p.text
				if p.ptype != "" {
					is_types[p.ptype] = true
				}
				param_list[i] = pair{
					k: p.name,
					v: strings.TrimSpace(text[:len(text)-len(p.name)]),
				}
			}
			info := command_info{
				rettype: strings.TrimSpace(text[:len(text)-len(c.proto.name)]),
				params:  param_list,
			}
			commands_map[c.proto.name] = info
		}
	}
	fmt.Println(template[0])
	fmt.Println(gen_c_def_type(is_types, api, registry.types))
	fmt.Println(template[1])
	for k, v := range commands_map {
		fmt.Println(gen_c_def_command(k, v))
	}
	fmt.Println(template[2])
	for k, _ := range commands_map {
		fmt.Println(gen_c_init_command(k))
	}
	fmt.Println(template[3])
	return nil
}

func main() {
	var argin string
	var argout string
	flag.StringVar(&argin, "i", "gl.xml", "input path of gl.xml")
	flag.StringVar(&argout, "o", ".", "output path of gl.go")
	flag.Parse()
	if !flag.Parsed() || flag.NArg() != 0 {
		panic("miss arg")
	}
	outpath, err := filepath.Abs(argout)
	if err != nil {
		panic(err)
	}
	if err := convert(argin, "", "3.2", outpath); err != nil {
		panic(err)
	}
}

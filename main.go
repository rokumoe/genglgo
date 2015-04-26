package main

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

var template = []string{
	`package gl
`,
	`
import (
	"unsafe"
)
`,
	`
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
import "C"
`,
	`
const (
`,
	`)
`,
	`
type (
`,
	`)
`,
	`
func Init() int {
	return int(C.gl_init())
}
`,
}

type pair struct {
	name  string
	type_ string
}

type type_info struct {
	name string
	text string
}

type command_info struct {
	params  []pair
	rettype string
}

func is_same_api(a string, b string) bool {
	return a == b || (a == "" && b == "gl") || (a == "gl" && b == "")
}

func kill_gl(s string) string {
	var prefixes = [...]string{
		"GL_",
		"GL",
		"gl",
	}
	for i := 0; i < len(prefixes); i++ {
		if strings.HasPrefix(s, prefixes[i]) && len(s) > len((prefixes[i])) {
			b := []byte(s[len(prefixes[i]):])
			if !unicode.IsDigit(rune(b[0])) {
				b[0] = byte(unicode.ToUpper(rune(b[0])))
				s = string(b)
			} else {
				s = "GL_" + string(b)
			}
		}
	}
	return s
}

var map_go_ptype = map[string]string{
	"Enum":     "uint32",
	"Boolean":  "uint8",
	"Bitfield": "uint32",
	"VoidPtr":  "unsafe.Pointer",
	"Byte":     "int8",
	"Short":    "int16",
	"Int":      "int32",
	"Ubyte":    "uint8",
	"Ushort":   "uint16",
	"Uint":     "uint32",
	"Sizei":    "int32",
	"Float":    "float32",
	"Clampf":   "float32",
	"Double":   "float64",
	"Clampd":   "float64",
	"Char":     "int8",
	"Half":     "uint16",
	"Intptr":   "uintptr",
	"Sizeiptr": "uintptr",
	"Int64":    "int64",
	"Uint64":   "uint64",
	"Sync":     "unsafe.Pointer",
	"Fixed":    "int32",
}

func gen_c_def_type(types []type_info) string {
	s := "\n"
	for _, t := range types {
		b := 0
		for i := 0; i <= len(t.text); i++ {
			if i == len(t.text) || t.text[i] == '\n' {
				s += "//" + t.text[b:i] + "\n"
				b = i + 1
			}
		}
	}
	return s + "import \"C\"\n\n"
}

func gen_c_def_command(command string, info command_info) string {
	var (
		params    string
		paramargs string
	)
	for _, p := range info.params {
		if params != "" {
			params += ", "
		}
		if paramargs != "" {
			paramargs += ", "
		}
		params += p.type_ + " " + p.name
		paramargs += p.name
	}
	def := "GLGO_COMMAND_DECL(" + info.rettype + ", " + command + ", " + params + ")\n"
	if info.rettype == "void" {
		def += "GLGO_COMMAND_0RET"
	} else {
		def += "GLGO_COMMAND_RET"
	}
	def += "(" + command + ", " + paramargs + ")\n"
	return def
}

func gen_c_init_command(command string) string {
	return `	if (GLGO_COMMAND_GETPROC(` + command + `) == NULL) {
		return __LINE__;
	}
`
}

var list_go_kw = [...]string{
	"type",
	"map",
	"func",
	"range",
	"cap",
	"string",
}

func save_gokw(w string) string {
	for _, k := range list_go_kw {
		if k == w {
			return w + "_"
		}
	}
	return w
}

func map_gotype(t string) string {
	if strings.Contains(t, "void *") {
		return "VoidPtr"
	}
	r := ""
	t = strings.Replace(t, "*", " * ", -1)
	for _, s := range strings.Split(t, " ") {
		if s == "*" {
			r = "*" + r
		} else if strings.HasPrefix(s, "GL") {
			r += kill_gl(s)
		}
	}
	return r
}

func gen_go_def_command(command string, info command_info) string {
	params := ""
	paramargs := ""
	for _, p := range info.params {
		name := save_gokw(p.name)
		if params != "" {
			params += ", "
		}
		params += name + " " + map_gotype(p.type_)
		if paramargs != "" {
			paramargs += ", "
		}
		paramargs += name
	}
	s := "func " + kill_gl(command) + "(" + params + ") "
	if info.rettype != "void" {
		rettype := map_gotype(info.rettype)
		s += rettype + " {\n"
		s += "\treturn " + rettype + "(C." + command + "(" + paramargs + "))\n"
	} else {
		s += " {\n"
		s += "\tC." + command + "(" + paramargs + ")\n"
	}
	s += "}\n"
	return s
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
					name:  p.name,
					type_: strings.TrimSpace(text[:len(text)-len(p.name)]),
				}
			}
			info := command_info{
				rettype: strings.TrimSpace(text[:len(text)-len(c.proto.name)]),
				params:  param_list,
			}
			commands_map[c.proto.name] = info
		}
	}
	for _, t := range registry.types.type_ {
		if is_types[t.name] && is_same_api(api, t.api) {
			if t.requires != "" {
				is_types[t.requires] = true
			}
		}
	}
	types_list := make([]type_info, 0, len(is_types))
	for _, t := range registry.types.type_ {
		if is_types[t.name] && is_same_api(api, t.api) {
			types_list = append(types_list, type_info{
				name: t.name,
				text: t.text,
			})
		}
	}
	f, err := os.OpenFile(glgo, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(template[0])
	f.WriteString(template[2])
	f.WriteString(gen_c_def_type(types_list))
	f.WriteString(template[3])
	for k, v := range commands_map {
		f.WriteString(gen_c_def_command(k, v))
	}
	f.WriteString(template[4])
	for k, _ := range commands_map {
		f.WriteString(gen_c_init_command(k))
	}
	f.WriteString(template[5])
	f.WriteString(template[6])
	for k, v := range enums_map {
		f.WriteString("\t" + kill_gl(k) + " = " + v + "\n")
	}
	f.WriteString(template[7])
	f.WriteString(template[8])
	for _, t := range types_list {
		if !strings.HasPrefix(t.name, "GL") {
			continue
		}
		gotype := kill_gl(t.name)
		if raw := map_go_ptype[gotype]; raw != "" {
			f.WriteString("\t" + gotype + " " + raw + "\n")
		}
	}
	f.WriteString(template[9])
	for command, info := range commands_map {
		f.WriteString(gen_go_def_command(command, info))
	}
	f.WriteString(template[10])
	return nil
}

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
	if err := convert(argin, argapi, argver, outpath); err != nil {
		panic(err)
	}
}

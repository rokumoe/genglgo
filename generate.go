package main

import (
	"os"
	"strconv"
	"strings"
	"unicode"
)

var templates = []string{
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
`,
	`
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

#ifndef GLGO_APIENTRYP
#define GLGO_APIENTRYP APIENTRY *
#endif

#ifdef far
#undef far
#endif
#ifdef near
#undef near
#endif

#define GLGO_COMMAND_DECL(return_type, command, ...) \
typedef return_type (GLGO_APIENTRYP _glgo_t_##command)(__VA_ARGS__);\
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
	`
/*
`,
	`
int gl_init()
{
`,
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
func Init() int {
	return int(C.gl_init())
}
`,
}

type param_info struct {
	name  string
	ptype string
}

type type_info struct {
	name string
	text string
}

type command_info struct {
	params  []param_info
	rettype string
}

var gl_prefix_list = [...]string{
	"GL_",
	"gl",
}

var go_rawtype_map = map[string]string{
	"GLenum":     "uint",
	"GLboolean":  "uint8",
	"GLbitfield": "uint32",
	"GLbyte":     "int8",
	"GLshort":    "int16",
	"GLint":      "int",
	"GLubyte":    "uint8",
	"GLushort":   "uint16",
	"GLuint":     "uint",
	"GLsizei":    "int",
	"GLfloat":    "float32",
	"GLclampf":   "float32",
	"GLdouble":   "float64",
	"GLclampd":   "float64",
	"GLchar":     "int8",
	"GLhalf":     "uint16",
	"GLintptr":   "uintptr",
	"GLsizeiptr": "uintptr",
	"GLint64":    "int64",
	"GLuint64":   "uint64",
	"GLsync":     "unsafe.Pointer",
	"GLfixed":    "int32",
}

var kw_list = [...]string{
	"type",
	"map",
	"func",
	"range",
	"cap",
	"string",
}

var (
	cgotype_map map[string]string
	gotype_map  map[string]string
)

func is_same_api(a string, b string) bool {
	return a == b || (a == "" && b == "gl") || (a == "gl" && b == "")
}

func kill_gl(s string) string {
	for i := 0; i < len(gl_prefix_list); i++ {
		if strings.HasPrefix(s, gl_prefix_list[i]) && len(s) > len((gl_prefix_list[i])) {
			b := []byte(s[len(gl_prefix_list[i]):])
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

func comment_text(t string) string {
	return "//" + strings.Replace(t, "\n", "\n//", -1)
}

func map_cgotype(t string) string {
	r := ""
	p := 0
	void := false
	for _, s := range strings.Split(strings.Replace(t, "*", " * ", -1), " ") {
		if s == "void" || s == "GLvoid" {
			void = true
		} else if s == "*" {
			p++
		} else if strings.HasPrefix(s, "GL") {
			r = "C." + s
		}
	}
	if void && p > 0 {
		r = "unsafe.Pointer"
		p--
	}
	return strings.Repeat("*", p) + r
}

func map_gotype(t string) string {
	r := ""
	p := 0
	void := false
	for _, s := range strings.Split(strings.Replace(t, "*", " * ", -1), " ") {
		if s == "void" || s == "GLvoid" {
			void = true
		} else if s == "*" {
			p++
		} else if strings.HasPrefix(s, "GL") {
			r = go_rawtype_map[s]
		}
	}
	if void && p > 0 {
		r = "unsafe.Pointer"
		p--
	}
	return strings.Repeat("*", p) + r
}

func gen_c_def_type(types []type_info) string {
	s := "\n"
	s += "//#ifndef __gl_h_\n"
	for _, t := range types {
		s += comment_text(t.text) + "\n"
	}
	s += "//#endif\n"
	return s + "import \"C\"\n"
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
		params += p.ptype + " " + p.name
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

func save_go_kw(w string) string {
	for _, k := range kw_list {
		if k == w {
			return w + "_"
		}
	}
	return w
}

func gen_go_func_command(command string, info command_info) string {
	params := ""
	paramargs := ""
	for i, p := range info.params {
		name := save_go_kw(p.name)
		if params != "" {
			params += ", "
		}
		params += name
		gotype := gotype_map[p.ptype]
		nexttype := ""
		if i < len(info.params)-1 {
			nexttype = gotype_map[info.params[i+1].ptype]
		}
		if gotype != nexttype {
			params += " " + gotype
		}
		if paramargs != "" {
			paramargs += ", "
		}
		cgotype := cgotype_map[p.ptype]
		if strings.HasPrefix(cgotype, "*") {
			cgotype = "(" + cgotype + ")"
			name = "unsafe.Pointer(" + name + ")"
		}
		paramargs += cgotype + "(" + name + ")"
	}
	s := "\n"
	s += "func " + kill_gl(command) + "(" + params + ") "
	if info.rettype != "void" {
		rettype := gotype_map[info.rettype]
		s += rettype + " {\n"
		if strings.HasPrefix(rettype, "*") {
			rettype = "(" + rettype + ")"
		}
		s += "\treturn " + rettype + "(C." + command + "(" + paramargs + "))\n"
	} else {
		s += "{\n"
		s += "\tC." + command + "(" + paramargs + ")\n"
	}
	s += "}\n"
	return s
}

func generate(glxml string, api string, number string, glgo string) error {
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
	cgotype_map = make(map[string]string)
	gotype_map = make(map[string]string)
	max_enums_len := 0
	for _, enums := range registry.enums {
		for _, e := range enums.enum {
			if is_enums[e.name] {
				name := kill_gl(e.name)
				enums_map[name] = e.value
				if len(name) > max_enums_len {
					max_enums_len = len(name)
				}
			}
		}
	}
	for _, c := range registry.commands.command {
		if is_commands[c.proto.name] {
			text := c.proto.text
			if c.proto.ptype != "" {
				is_types[c.proto.ptype] = true
			}
			param_list := make([]param_info, len(c.param))
			for i, p := range c.param {
				text := p.text
				if p.ptype != "" {
					is_types[p.ptype] = true
				}
				ptype := strings.TrimSpace(text[:len(text)-len(p.name)])
				param_list[i] = param_info{
					name:  p.name,
					ptype: ptype,
				}
				gotype_map[ptype] = map_gotype(ptype)
				cgotype_map[ptype] = map_cgotype(ptype)
			}
			rettype := strings.TrimSpace(text[:len(text)-len(c.proto.name)])
			info := command_info{
				rettype: rettype,
				params:  param_list,
			}
			gotype_map[rettype] = map_gotype(rettype)
			cgotype_map[rettype] = map_cgotype(rettype)
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
	ctypes_list := make([]type_info, 0, len(is_types))
	for _, t := range registry.types.type_ {
		if is_types[t.name] && is_same_api(api, t.api) {
			ctypes_list = append(ctypes_list, type_info{
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
	f.WriteString(templates[0])
	f.WriteString(templates[1])
	f.WriteString(templates[2])
	f.WriteString(gen_c_def_type(ctypes_list))
	f.WriteString(templates[3])
	f.WriteString(templates[4])
	for k, v := range commands_map {
		f.WriteString(gen_c_def_command(k, v))
	}
	f.WriteString(templates[5])
	for k, _ := range commands_map {
		f.WriteString(gen_c_init_command(k))
	}
	f.WriteString(templates[6])
	f.WriteString("\nconst GL_VERSION_NUMBER = \"" + number + "\"\n")
	f.WriteString(templates[7])
	for k, v := range enums_map {
		f.WriteString("\t" + k + strings.Repeat(" ", max_enums_len-len(k)) + " = " + v + "\n")
	}
	f.WriteString(templates[8])
	for command, info := range commands_map {
		f.WriteString(gen_go_func_command(command, info))
	}
	f.WriteString(templates[9])
	return nil
}

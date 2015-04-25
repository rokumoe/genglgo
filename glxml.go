package main

import (
	"sync"
)

type registry_commands_command_glx struct {
	type_   string
	opcode  string
	name    string
	comment string
}

type registry_commands_command_param struct {
	group string
	len_  string
	name  string
	ptype string
	text  string
}

type registry_commands_command struct {
	proto struct {
		group string
		name  string
		ptype string
		text  string
	}
	alias struct {
		name string
	}
	vecequiv struct {
		name string
	}
	param []registry_commands_command_param
	glx   []registry_commands_command_glx
}

type registry_commands struct {
	namespace string
	command   []registry_commands_command
}

type registry_enums_enum struct {
	alias   string
	api     string
	comment string
	name    string
	type_   string
	value   string
}

type registry_enums_unused struct {
	comment string
	end     string
	start   string
	vendor  string
}

type registry_enums struct {
	comment   string
	end       string
	group     string
	namespace string
	start     string
	type_     string
	vendor    string
	enum      []registry_enums_enum
	unused    []registry_enums_unused
}

type registry_extensions_extension_require_command struct {
	comment string
	name    string
}

type registry_extensions_extension_require_enum struct {
	name string
}

type registry_extensions_extension_require_type struct {
	name string
}

type registry_extensions_extension_require struct {
	api     string
	comment string
	profile string
	command []registry_extensions_extension_require_command
	enum    []registry_extensions_extension_require_enum
	type_   []registry_extensions_extension_require_type
}

type registry_extensions_extension struct {
	comment   string
	name      string
	supported string
	require   []registry_extensions_extension_require
}

type registry_extensions struct {
	extension []registry_extensions_extension
}

type registry_feature_remove_command struct {
	name string
}

type registry_feature_remove_enum struct {
	name string
}

type registry_feature_remove struct {
	comment string
	profile string
	command []registry_feature_remove_command
	enum    []registry_feature_remove_enum
}

type registry_feature_require_command struct {
	name string
}

type registry_feature_require_enum struct {
	name    string
	comment string
}

type registry_feature_require_type struct {
	name    string
	comment string
}
type registry_feature_require struct {
	comment string
	profile string
	command []registry_feature_require_command
	enum    []registry_feature_require_enum
	type_   []registry_feature_require_type
}

type registry_feature struct {
	api     string
	name    string
	number  string
	require []registry_feature_require
	remove  []registry_feature_remove
}

type registry_groups_group_enum struct {
	name string
}

type registry_groups_group struct {
	name    string
	comment string
	enum    []registry_groups_group_enum
}

type registry_groups struct {
	group []registry_groups_group
}

type registry_types_type struct {
	api      string
	comment  string
	name     string
	requires string
	text     string
}

type registry_types struct {
	type_ []registry_types_type
}

type glxml_registry struct {
	comment    string
	types      registry_types
	groups     registry_groups
	enums      []registry_enums
	commands   registry_commands
	feature    []registry_feature
	extensions registry_extensions
}

//xpath:/registry/comment
func glxml_parse_comment(node *xnode, registry *glxml_registry, l *sync.Mutex) {
	l.Lock()
	registry.comment = node.children[0].value.(string)
	l.Unlock()
}

//xpath:/registry/types
func glxml_parse_types(node *xnode, registry *glxml_registry, l *sync.Mutex) {
	for _, e := range node.elements("type") {
		var type_ registry_types_type
		type_.name = e.attr("name")
		if type_.name == "" {
			type_.name = e.elements("name")[0].text()
		}
		type_.comment = e.attr("comment")
		type_.requires = e.attr("requires")
		type_.api = e.attr("api")
		type_.text = e.text()
		registry.types.type_ = append(registry.types.type_, type_)
	}
}

//xpath:/registry/groups
func glxml_parse_groups(node *xnode, registry *glxml_registry, l *sync.Mutex) {
	for _, e := range node.elements("group") {
		var group registry_groups_group
		group.name = e.attr("name")
		group.comment = e.attr("comment")
		for _, enum := range e.elements("enum") {
			group.enum = append(group.enum, registry_groups_group_enum{name: enum.attr("name")})
		}
		l.Lock()
		registry.groups.group = append(registry.groups.group, group)
		l.Unlock()
	}
}

//xpath:/registry/enums
func glxml_parse_enums(node *xnode, registry *glxml_registry, l *sync.Mutex) {
	var enums registry_enums
	enums.namespace = node.attr("namespace")
	enums.group = node.attr("group")
	enums.type_ = node.attr("type")
	enums.comment = node.attr("comment")
	enums.vendor = node.attr("vendor")
	enums.start = node.attr("start")
	enums.end = node.attr("end")
	for _, e := range node.elements("enum") {
		var enum registry_enums_enum
		enum.value = e.attr("value")
		enum.name = e.attr("name")
		enum.comment = e.attr("comment")
		enum.type_ = e.attr("type")
		enum.alias = e.attr("alias")
		enum.api = e.attr("api")
		enums.enum = append(enums.enum, enum)
	}
	for _, e := range node.elements("unused") {
		var unused registry_enums_unused
		unused.start = e.attr("start")
		unused.end = e.attr("end")
		unused.vendor = e.attr("vendor")
		unused.comment = e.attr("comment")
		enums.unused = append(enums.unused, unused)
	}
	l.Lock()
	registry.enums = append(registry.enums, enums)
	l.Unlock()
}

//xpath:/registry/commands
func glxml_parse_commands(node *xnode, registry *glxml_registry, l *sync.Mutex) {
	l.Lock()
	registry.commands.namespace = node.attr("namespace")
	l.Unlock()
	for _, e := range node.elements("command") {
		var command registry_commands_command
		proto := e.elements("proto")[0]
		command.proto.group = proto.attr("group")
		command.proto.name = proto.elements("name")[0].text()
		if ptype := proto.elements("ptype"); len(ptype) > 0 {
			command.proto.ptype = ptype[0].text()
		}
		command.proto.text = proto.text()
		for _, eparam := range e.elements("param") {
			var param registry_commands_command_param
			param.group = eparam.attr("group")
			param.len_ = eparam.attr("len")
			param.name = eparam.elements("name")[0].text()
			ptype := eparam.elements("ptype")
			if len(ptype) > 0 {
				param.ptype = ptype[0].text()
			}
			param.text = eparam.text()
			command.param = append(command.param, param)
		}
		for _, eglx := range e.elements("glx") {
			var glx registry_commands_command_glx
			glx.type_ = eglx.attr("type")
			glx.opcode = eglx.attr("opcode")
			glx.name = eglx.attr("name")
			glx.comment = eglx.attr("comment")
			command.glx = append(command.glx, glx)
		}
		if ealias := e.elements("alias"); len(ealias) == 1 {
			command.alias.name = ealias[0].attr("name")
		}
		if evecequiv := e.elements("vecequiv"); len(evecequiv) == 1 {
			command.vecequiv.name = evecequiv[0].attr("name")
		}
		l.Lock()
		registry.commands.command = append(registry.commands.command, command)
		l.Unlock()
	}
}

//xpath:/registry/feature
func glxml_parse_feature(node *xnode, registry *glxml_registry, l *sync.Mutex) {
	var feature registry_feature
	feature.api = node.attr("api")
	feature.name = node.attr("name")
	feature.number = node.attr("number")
	for _, e := range node.elements("require") {
		var require registry_feature_require
		require.comment = e.attr("comment")
		require.profile = e.attr("profile")
		for _, eenum := range e.elements("enum") {
			require.enum = append(require.enum, registry_feature_require_enum{
				name:    eenum.attr("name"),
				comment: eenum.attr("comment"),
			})
		}
		for _, ecommand := range e.elements("command") {
			require.command = append(require.command, registry_feature_require_command{
				name: ecommand.attr("name"),
			})
		}
		for _, etype := range e.elements("type") {
			require.type_ = append(require.type_, registry_feature_require_type{
				name:    etype.attr("name"),
				comment: etype.attr("comment"),
			})
		}
		feature.require = append(feature.require, require)
	}
	for _, e := range node.elements("remove") {
		var remove registry_feature_remove
		remove.comment = e.attr("comment")
		remove.profile = e.attr("profile")
		for _, ecommand := range e.elements("command") {
			remove.command = append(remove.command, registry_feature_remove_command{
				name: ecommand.attr("name"),
			})
		}
		for _, eenum := range e.elements("enum") {
			remove.enum = append(remove.enum, registry_feature_remove_enum{
				name: eenum.attr("name"),
			})
		}
		feature.remove = append(feature.remove, remove)
	}
	l.Lock()
	registry.feature = append(registry.feature, feature)
	l.Unlock()
}

//xpath:/registry/extensions
func glxml_parse_extensions(node *xnode, registry *glxml_registry, l *sync.Mutex) {
	for _, e := range node.elements("extension") {
		var extension registry_extensions_extension
		extension.name = e.attr("name")
		extension.comment = e.attr("comment")
		extension.supported = e.attr("supported")
		for _, erequire := range e.elements("require") {
			var require registry_extensions_extension_require
			require.api = erequire.attr("api")
			require.comment = erequire.attr("comment")
			require.profile = erequire.attr("profile")
			for _, ecommand := range erequire.elements("command") {
				require.command = append(require.command, registry_extensions_extension_require_command{
					name:    ecommand.attr("name"),
					comment: ecommand.attr("comment"),
				})
			}
			for _, eenum := range erequire.elements("enum") {
				require.enum = append(require.enum, registry_extensions_extension_require_enum{
					name: eenum.attr("name"),
				})
			}
			for _, etype := range erequire.elements("type") {
				require.type_ = append(require.type_, registry_extensions_extension_require_type{
					name: etype.attr("name"),
				})
			}
			extension.require = append(extension.require, require)
		}
		l.Lock()
		registry.extensions.extension = append(registry.extensions.extension, extension)
		l.Unlock()
	}
}

func load_glxml(glxml string) (*glxml_registry, error) {
	root, err := loadxml(glxml, false)
	if err != nil {
		return nil, err
	}
	parse_handler := map[string]func(node *xnode, registry *glxml_registry, l *sync.Mutex){
		"comment":    glxml_parse_comment,
		"types":      glxml_parse_types,
		"groups":     glxml_parse_groups,
		"enums":      glxml_parse_enums,
		"commands":   glxml_parse_commands,
		"feature":    glxml_parse_feature,
		"extensions": glxml_parse_extensions,
	}
	l := sync.Mutex{}
	wg := sync.WaitGroup{}
	registry := new(glxml_registry)
	node := root.elements("registry")
	for _, child := range node[0].children {
		if child.xtype != xml_element {
			continue
		}
		wg.Add(1)
		go func(c *xnode) {
			parse_handler[c.name](c, registry, &l)
			wg.Done()
		}(child)
	}
	wg.Wait()
	return registry, nil
}

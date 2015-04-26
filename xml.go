package main

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"reflect"
	"unicode"
)

type xmltype int

const (
	xml_element xmltype = iota
	xml_chardata
	xml_comment
	xml_directive
	xml_procinst
)

type xnode struct {
	parent   *xnode
	children []*xnode
	xtype    xmltype
	name     string
	value    interface{}
}

func (node *xnode) add(xtype xmltype, name string, value interface{}) *xnode {
	n := &xnode{
		parent:   node,
		children: nil,
		xtype:    xtype,
		name:     name,
		value:    value,
	}
	node.children = append(node.children, n)
	return n
}

func (node *xnode) find(xtype xmltype, name string) []*xnode {
	var nodes []*xnode = nil
	for _, child := range node.children {
		if child.xtype == xtype && child.name == name {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (node *xnode) elements(name string) []*xnode {
	return node.find(xml_element, name)
}

func (node *xnode) attr(name string) string {
	attrs, ok := node.value.(map[string]string)
	if !ok {
		return ""
	}
	return attrs[name]
}

func (node *xnode) text() string {
	switch node.xtype {
	case xml_chardata:
		return node.value.(string)
	case xml_element:
		s := ""
		for _, c := range node.children {
			s += c.text()
		}
		return s
	}
	return ""
}

func iswhite(s string) bool {
	for _, c := range s {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

func loadxml(path string, nowhite bool) (*xnode, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := xml.NewDecoder(f)
	root := new(xnode)
	cur := root
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch t := token.(type) {
		case xml.StartElement:
			var value interface{} = nil
			if len(t.Attr) != 0 {
				attrs := make(map[string]string, len(t.Attr))
				for _, a := range t.Attr {
					attrs[a.Name.Local] = a.Value
				}
				value = attrs
			}
			cur = cur.add(xml_element, t.Name.Local, value)
		case xml.EndElement:
			if cur.name != t.Name.Local {
				return nil, errors.New("bad element: " + cur.name)
			}
			cur = cur.parent
		case xml.CharData:
			data := string(t)
			if nowhite && iswhite(data) {
				break
			}
			cur.add(xml_chardata, "", data)
		case xml.Comment:
			cur.add(xml_comment, "", string(t))
		case xml.Directive:
			cur.add(xml_directive, "", string(t))
		case xml.ProcInst:
			cur.add(xml_procinst, t.Target, string(t.Inst))
		default:
			return nil, errors.New("bad type: " + reflect.TypeOf(t).Name())
		}
	}
	return root, nil
}

package withcheck

import (
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/knsh14/templateutil"
)

var (
	// ErrNotFound is error for variable is not used in with statement
	ErrNotFound = fmt.Errorf("not found")
)

// Check returns error if variable is unused in with statement or some error
func Check(tmpl *template.Template) error {
	var err error
	templateutil.Inspect(tmpl.Tree.Root, func(node parse.Node) bool {
		if err != nil {
			return false
		}
		if n, ok := node.(*parse.WithNode); ok {
			v := getVariable(n.Pipe)
			if len(v) == 0 {
				return false
			}
			e := checkVariable(n.List, v)
			if e != nil {
				err = e
			}
			return false
		}
		return true
	})
	return err
}

func checkVariable(list *parse.ListNode, variables []string) error {
	found := false
	for _, target := range variables {
		templateutil.Inspect(list, func(node parse.Node) bool {
			f := false
			switch n := node.(type) {
			case *parse.FieldNode:
				v := "." + strings.Join(n.Ident, ".")
				f = strings.HasPrefix(v, target)
			case *parse.IdentifierNode:
				f = strings.HasPrefix(n.Ident, target)
			case *parse.VariableNode:
				v := "." + strings.Join(n.Ident, ".")
				f = strings.HasPrefix(v, target)
			case *parse.ChainNode:
				v := "." + strings.Join(n.Field, ".")
				f = strings.HasPrefix(v, target)
			case *parse.DotNode:
				f = target == "."
			case *parse.TemplateNode:
				f = n.Name == target
			}
			found = found || f
			return true
		})
	}
	if found {
		return nil
	}
	return ErrNotFound
}

func getVariable(n *parse.PipeNode) []string {
	var names []string
	if len(n.Decl) > 0 {
		for _, v := range n.Decl {
			names = append(names, v.Ident...)
		}
		return names
	}
	if len(n.Cmds) > 0 {
		for _, cmd := range n.Cmds {
			for _, arg := range cmd.Args {
				switch node := arg.(type) {
				case *parse.FieldNode:
					names = append(names, "."+strings.Join(node.Ident, "."))
				case *parse.DotNode:
					names = append(names, ".")
				}
			}
		}
		return names
	}
	return nil
}

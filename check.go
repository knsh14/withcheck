package withcheck

import (
	"errors"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/knsh14/templateutil"
)

var (
	// ErrNotFound is error for variable is not used in with statement
	ErrNotFound = errors.New("not found")
	// ErrTooManyVariables is error for is not used in with statement
	ErrTooManyVariables = errors.New("too many variables")
	// ErrInvalid is error for variable is not used in with statement
	ErrInvalid = errors.New("invalid node")
)

// Check returns error if variable is unused in with statement or some error
func Check(tmpl *template.Template) error {
	var err error
	templateutil.Inspect(tmpl.Tree.Root, func(node parse.Node) bool {
		if err != nil {
			return false
		}
		if n, ok := node.(*parse.WithNode); ok {
			v, e := getVariable(n.Pipe)
			if err != nil {
				err = e
				return false
			}
			e = checkVariable(n.List, v)
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
	var found bool
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
				v := strings.Join(n.Ident, ".")
				f = strings.HasPrefix(v, target)
			case *parse.ChainNode:
				v := "." + strings.Join(n.Field, ".")
				f = strings.HasPrefix(v, target)
			case *parse.DotNode:
				f = target == "."
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

func getVariable(n *parse.PipeNode) ([]string, error) {
	if len(n.Decl) > 0 {
		if len(n.Decl) > 1 {
			return nil, ErrTooManyVariables
		}
		return []string{n.Decl[0].Ident[0], "."}, nil
	}
	if len(n.Cmds) > 0 {
		args := n.Cmds[0].Args
		if len(args) > 1 {
			return nil, ErrTooManyVariables
		}
		if _, ok := args[0].(*parse.FieldNode); ok {
			return []string{"."}, nil
		}
		if _, ok := args[0].(*parse.DotNode); ok {
			return []string{"."}, nil
		}
		return nil, ErrInvalid
	}
	return nil, ErrNotFound
}

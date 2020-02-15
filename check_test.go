package withcheck

import (
	"fmt"
	"testing"
	"text/template"
	"text/template/parse"

	"github.com/knsh14/templateutil"
)

func TestCheck(t *testing.T) {
	testcases := []struct {
		title  string
		input  string
		expect error
	}{
		{
			title:  "field",
			input:  `{{ with .Foo}}{{.}}{{end}}`,
			expect: nil,
		},
		{
			title:  "simple not found",
			input:  `{{ with .Foo}}{{end}}`,
			expect: ErrNotFound,
		},
		{
			title:  "variable",
			input:  `{{ with $x := "a" }}{{ $x }}{{end}}`,
			expect: nil,
		},
		{
			title:  "variable",
			input:  `{{ with $x := "b" }}{{ . }}{{end}}`,
			expect: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.title, func(t *testing.T) {
			tpl, err := template.New("test").Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			err = Check(tpl)
			if err != tt.expect {
				t.Fatal(err)
			}
		})
	}
}

func TestCheckVariables(t *testing.T) {
	testcases := []struct {
		input  string
		target []string
		expect error
	}{
		{
			input:  `{{.Foo}}`,
			target: []string{"."},
			expect: nil,
		},
		{
			input:  `hello world`,
			target: []string{"."},
			expect: ErrNotFound,
		},
		{
			input:  `{{ . }}`,
			target: []string{"."},
			expect: nil,
		},
		{
			input:  `{{ .Foo }}`,
			target: []string{"."},
			expect: nil,
		},
		{
			input:  `{{ $x := "a"}}{{ $x }}`,
			target: []string{"$x", "."},
			expect: nil,
		},
		{
			input:  `{{ .Foo }}`,
			target: []string{"$x"},
			expect: ErrNotFound,
		},
		{
			input:  `{{ $x := "b" }}{{ $x.Y }}`,
			target: []string{"$x"},
			expect: nil,
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tpl, err := template.New("test").Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			err = checkVariable(tpl.Tree.Root, tt.target)
			if err != tt.expect {
				t.Errorf("unexpected result, expected:%v, actual:%v", tt.expect, err)
			}
		})
	}
}

func TestGetVariable(t *testing.T) {
	testcases := []struct {
		input    string
		expected []string
		err      error
	}{
		{
			input:    `{{ with . }}{{ .Foo }}{{ end }}`,
			expected: []string{"."},
		},
		{
			input:    `{{ with .Foo }}{{ .Foo }}{{ end }}`,
			expected: []string{"."},
		},
		{
			input:    `{{ with .Foo | println "Bar" }}{{ .Foo }}{{ end }}`,
			expected: []string{"."},
		},
		{
			input:    `{{ with .Foo .Bar | println "Bar" }}{{ .Foo }}{{ end }}`,
			expected: nil,
			err:      ErrTooManyVariables,
		},
		{
			input:    `{{ with $x := "huga" }}{{ $x }}{{ end }}`,
			expected: []string{"$x", "."},
		},
		{
			input:    `{{ with $x := "huga" | println "hoge" }}{{ $x }}{{ end }}`,
			expected: []string{"$x", "."},
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tpl, err := template.New("test").Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			templateutil.Inspect(tpl.Tree.Root, func(n parse.Node) bool {
				if node, ok := n.(*parse.WithNode); ok {
					result, err := getVariable(node.Pipe)
					if err != tt.err {
						t.Fatalf("error is unexpected. Actual:%s, Expected:%s", err, tt.err)
					}
					if len(result) != len(tt.expected) {
						t.Fatalf("result num is unexpected. Actual:%d, Expected:%d", len(result), len(tt.expected))
					}
					for i := range result {
						if result[i] != tt.expected[i] {
							t.Fatalf("result is unexpected. Actual:%s, Expected:%s", result[i], tt.expected[i])
						}
					}
				}
				return true
			})
		})
	}
}

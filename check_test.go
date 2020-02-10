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
			title:  "simple",
			input:  `{{ with .Foo}}{{.Foo}}{{end}}`,
			expect: nil,
		},
		{
			title:  "simple not found",
			input:  `{{ with .Foo}}{{.Bar}}{{end}}`,
			expect: ErrNotFound,
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
			target: []string{".Foo"},
			expect: nil,
		},
		{
			input:  `{{.Bar}}`,
			target: []string{".Foo"},
			expect: ErrNotFound,
		},
		{
			input:  `{{ .Foo }}{{ .Bar }}`,
			target: []string{".Foo", ".Bar"},
			expect: nil,
		},
		{
			input:  `{{ .Bar }}`,
			target: []string{".Foo", ".Bar"},
			expect: nil,
		},
		{
			input:  `{{ . }}`,
			target: []string{"."},
			expect: nil,
		},
		{
			input:  `{{ template "hoge" }}`,
			target: []string{"hoge"},
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
	}{
		{
			input:    `{{ with . }}{{ .Foo }}{{ end }}`,
			expected: []string{"."},
		},
		{
			input:    `{{ with .Foo }}{{ .Foo }}{{ end }}`,
			expected: []string{".Foo"},
		},
		{
			input:    `{{ with .Foo | println "Bar" }}{{ .Foo }}{{ end }}`,
			expected: []string{".Foo"},
		},
		{
			input:    `{{ with .Foo .Bar | println "Bar" }}{{ .Foo }}{{ end }}`,
			expected: []string{".Foo", ".Bar"},
		},
		{
			input:    `{{ with $x := "huga" }}{{ $x }}{{ end }}`,
			expected: []string{"$x"},
		},
		{
			input:    `{{ with $x := "huga" | println "hoge" }}{{ $x }}{{ end }}`,
			expected: []string{"$x"},
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
					result := getVariable(node.Pipe)
					if len(result) != len(tt.expected) {
						t.Fatalf("result is not matched to expected, Expected:%+v, Actual:%+v", tt.expected, result)
					}
					for i := range result {
						if result[i] != tt.expected[i] {
							t.Errorf("value is not matched , Expected:%s, Actual:%s", tt.expected[i], result[i])
						}
					}
				}
				return true
			})
		})
	}
}

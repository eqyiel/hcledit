package hcledit_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"go.mercari.io/hcledit"
)

func TestCreate(t *testing.T) {
	cases := map[string]struct {
		input string
		query string
		value interface{}
		want  string
	}{
		"Attribute": {
			input: `
`,
			query: "attribute",
			value: "C",
			want: `
attribute = "C"
`,
		},

		"AttributeInBlock1": {
			input: `
block "label" {
}
`,
			query: "block.label.attribute",
			value: "C",
			want: `
block "label" {
  attribute = "C"
}
`,
		},

		"AttributeInBlock2": {
			input: `
block1 "label1" {
  block2 "label2" {
  }
}
`,
			query: "block1.label1.block2.label2.attribute",
			value: "C",
			want: `
block1 "label1" {
  block2 "label2" {
    attribute = "C"
  }
}
`,
		},

		"AttributeInBlock3": {
			input: `
block "label" "label1"{
}

block "label" "label2" {
}
`,
			query: "block.label.*.attribute",
			value: "C",
			want: `
block "label" "label1" {
  attribute = "C"
}

block "label" "label2" {
  attribute = "C"
}
`,
		},

		"AttributeInObject1": {
			input: `
object = {
}
`,
			query: "object.attribute",
			value: "C",
			want: `
object = {
  attribute = "C"
}
`,
		},

		"AttributeInObject2": {
			input: `
object = {
  attribute1 = "str1"
}
`,
			query: "object.attribute2",
			value: "C",
			want: `
object = {
  attribute1 = "str1"
  attribute2 = "C"
}
`,
		},

		"Block": {
			input: `
`,
			query: "block",
			value: hcledit.BlockVal("label1", "label2"),
			want: `
block "label1" "label2" {
}
`,
		},

		"Raw": {
			input: `
`,
			query: "object1",
			value: hcledit.RawVal(`{
  object2 = {
    attribute1 = "str1"
  }
}`),
			want: `
object1 = {
  object2 = {
    attribute1 = "str1"
  }
}
`,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			editor, err := hcledit.Read(strings.NewReader(tc.input), "")
			if err != nil {
				t.Fatal(err)
			}
			if err := editor.Create(tc.query, tc.value); err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(tc.want, string(editor.Bytes()),
				cmpopts.AcyclicTransformer("multiline", func(s string) []string {
					return strings.Split(s, "\n")
				}),
			)
			if diff != "" {
				t.Errorf("Create() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRead(t *testing.T) {
	cases := map[string]struct {
		input string
		query string
		want  map[string]interface{}
	}{
		"Attribute": {
			input: `
attribute = "R"
`,
			query: "attribute",
			want: map[string]interface{}{
				"attribute": "R",
			},
		},

		// TODO(tcnksm): Now block and object is not supported to read.
		"Block": {
			input: `
block "label1" "label2" {
  attribute = "str"
}
`,
			query: "block",
			want:  map[string]interface{}{},
		},

		"AttributeInBlock1": {
			input: `
block "label1" "label2" {
  attribute = "R"
}
`,
			query: "block.label1.label2.attribute",
			want: map[string]interface{}{
				"block.label1.label2.attribute": "R",
			},
		},

		"AttributeInBlock2": {
			input: `
block1 "label1" "label2" {
  block2 "label3" "label4" {
    attribute = "R"
  }
}
`,
			query: "block1.label1.label2.block2.label3.label4.attribute",
			want: map[string]interface{}{
				"block1.label1.label2.block2.label3.label4.attribute": "R",
			},
		},

		"AttributeInBlock3": {
			input: `
block "label" "label1" {
  attribute = "R"
}

block "label" "label2" {
  attribute = "R"
}

`,
			query: "block.label.*.attribute",
			want: map[string]interface{}{
				"block.label.label1.attribute": "R",
				"block.label.label2.attribute": "R",
			},
		},

		"AttributeInObject1": {
			input: `
object = {
  attribute = "R"
}
`,
			query: "object.attribute",
			want: map[string]interface{}{
				"object.attribute": "R",
			},
		},
		"AttributeInObject2": {
			input: `
object1 = {
  object2 = {
    attribute = "R"
  }  
}
`,
			query: "object1.object2.attribute",
			want: map[string]interface{}{
				"object1.object2.attribute": "R",
			},
		},

		"TypeNumber": {
			input: `
attribute = 1
`,
			query: "attribute",
			want: map[string]interface{}{
				"attribute": 1,
			},
		},

		"TypeString": {
			input: `
attribute = "str"
`,
			query: "attribute",
			want: map[string]interface{}{
				"attribute": "str",
			},
		},

		"TypeBool1": {
			input: `
attribute = true
`,
			query: "attribute",
			want: map[string]interface{}{
				"attribute": true,
			},
		},

		"TypeBool2": {
			input: `
attribute = false
`,
			query: "attribute",
			want: map[string]interface{}{
				"attribute": false,
			},
		},

		"TypeStringList": {
			input: `
attribute = ["str1", "str2", "str3"]
`,
			query: "attribute",
			want: map[string]interface{}{
				"attribute": []string{"str1", "str2", "str3"},
			},
		},

		"TypeIntList": {
			input: `
attribute = [1, 2, 3]
`,
			query: "attribute",
			want: map[string]interface{}{
				"attribute": []int{1, 2, 3},
			},
		},

		"TypeBoolList": {
			input: `
attribute = [true, false, true]
`,
			query: "attribute",
			want: map[string]interface{}{
				"attribute": []bool{true, false, true},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			editor, err := hcledit.Read(strings.NewReader(tc.input), "")
			if err != nil {
				t.Fatal(err)
			}

			got, err := editor.Read(tc.query)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Read() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	cases := map[string]struct {
		input string
		query string
		value interface{}
		want  string
	}{
		"Attribute1": {
			input: `
attribute = "str"
`,
			query: "attribute",
			value: "U",
			want: `
attribute = "U"
`,
		},

		"Attribute2": {
			input: `
attribute1 = "str1"
attribute2 = "str2"
`,
			query: "attribute1",
			value: "U",
			want: `
attribute1 = "U"
attribute2 = "str2"
`,
		},

		"AttributeWithComment1": {
			input: `
// Comment
attribute = "str"
`,
			query: "attribute",
			value: "U",
			want: `
// Comment
attribute = "U"
`,
		},

		"AttributeWithComment2": {
			input: `
// Comment1
attribute1 = "str1"

// Comment2
attribute2 = "str2"
`,
			query: "attribute1",
			value: "U",
			want: `
// Comment1
attribute1 = "U"

// Comment2
attribute2 = "str2"
`,
		},

		"AttributeInBlock1": {
			input: `
block "label" {
  attribute = "str"
}
`,
			query: "block.label.attribute",
			value: "U",
			want: `
block "label" {
  attribute = "U"
}
`,
		},

		"AttributeInBlock2": {
			input: `
block "label1" "label2" {
  attribute = "str"
}
`,
			query: "block.label1.label2.attribute",
			value: "U",
			want: `
block "label1" "label2" {
  attribute = "U"
}
`,
		},

		"AttributeInBlock3": {
			input: `
block1 "label1" "label2" {
  block2 {
    attribute = "str"
  }
}
`,
			query: "block1.label1.label2.block2.attribute",
			value: "U",
			want: `
block1 "label1" "label2" {
  block2 {
    attribute = "U"
  }
}
`,
		},

		"AttributeInBlock4": {
			input: `
block1 "label1" "label2" {
  block2 "label3" "label4" {
    attribute = "str"
  }
}
`,
			query: "block1.label1.label2.block2.label3.label4.attribute",
			value: "U",
			want: `
block1 "label1" "label2" {
  block2 "label3" "label4" {
    attribute = "U"
  }
}
`,
		},

		"AttributeInObject1": {
			input: `
object = {
  attribute = "str"
}
`,
			query: "object.attribute",
			value: "U",
			want: `
object = {
  attribute = "U"
}
`,
		},

		"AttributeInObject2": {
			input: `
object1 = {
  object2 = {
    attribute = "str"
  }
}
`,
			query: "object1.object2.attribute",
			value: "U",
			want: `
object1 = {
  object2 = {
    attribute = "U"
  }
}
`,
		},

		"AttributeInObject3": {
			input: `
object = {
  attribute1 = "str1"
  attribute2 = "str2"
}
`,
			query: "object.attribute1",
			value: "U",
			want: `
object = {
  attribute1 = "U"
  attribute2 = "str2"
}
`,
		},

		"AttributeInObject4": {
			input: `
object1 = {
  attribute1 = "str1"
}

object2 = {
  attribute2 = "str2"
}
`,
			query: "object1.attribute1",
			value: "U",
			want: `
object1 = {
  attribute1 = "U"
}

object2 = {
  attribute2 = "str2"
}
`,
		},

		"TypeString": {
			input: `
attribute = "T"
`,
			query: "attribute",
			value: "str",
			want: `
attribute = "str"
`,
		},

		"TypeInt": {
			input: `
attribute = "T"
`,
			query: "attribute",
			value: 1,
			want: `
attribute = 1
`,
		},

		// TODO(tcnksm)
		"TypeFloat": {
			input: `
attribute = "T"
`,
			query: "attribute",
			value: 1.0,
			want: `
attribute = 1
`,
		},

		"TypeBool1": {
			input: `
attribute = "T"
`,
			query: "attribute",
			value: true,
			want: `
attribute = true
`,
		},

		"TypeBool2": {
			input: `
attribute = "T"
`,
			query: "attribute",
			value: false,
			want: `
attribute = false
`,
		},

		"TypeStringList": {
			input: `
attribute = "T"
`,
			query: "attribute",
			value: []string{"str1", "str2", "str3"},
			want: `
attribute = ["str1", "str2", "str3"]
`,
		},

		"TypeNubberList": {
			input: `
attribute = "T"
`,
			query: "attribute",
			value: []int{1, 2, 3},
			want: `
attribute = [1, 2, 3]
`,
		},

		"TypeBoolList": {
			input: `
attribute = "T"
`,
			query: "attribute",
			value: []bool{true, false, true},
			want: `
attribute = [true, false, true]
`,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			editor, err := hcledit.Read(strings.NewReader(tc.input), "")
			if err != nil {
				t.Fatal(err)
			}
			if err := editor.Update(tc.query, tc.value); err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(tc.want, string(editor.Bytes()),
				cmpopts.AcyclicTransformer("multiline", func(s string) []string {
					return strings.Split(s, "\n")
				}),
			)

			if diff != "" {
				t.Errorf("Update() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cases := map[string]struct {
		input string
		query string
		want  string
	}{
		"Block1": {
			input: `
block1 "label1" "label2" {
  block2 {
    attribute = "str1"
  }
}
`,
			query: "block1.label1.label2.block2",
			want: `
block1 "label1" "label2" {
}
`,
		},

		"Block2": {
			input: `
block1 "label1" "label2" {
  block2 {
    attribute = "str1"
  }
}
`,
			query: "block1",
			want: `
`,
		},

		"Block3": {
			input: `
block1 "label1" "label2" {
  block2 "label3" {
    attribute = "str1"
  }
}
`,
			query: "block1.label1.label2.block2",
			want: `
block1 "label1" "label2" {
}
`,
		},

		"Attribute": {
			input: `
attribute = "str1"
`,
			query: "attribute",
			want: `
`,
		},

		"AttributeInBlock1": {
			input: `
block "label" {
  attribute = "str1"
}
`,
			query: "block.label.attribute",
			want: `
block "label" {
}
`,
		},

		"AttributeInBlock2": {
			input: `
block "label1" "label2" {
  attribute = "str1"
}
`,
			query: "block.label1.label2.attribute",
			want: `
block "label1" "label2" {
}
`,
		},

		"AttributeInBlock3": {
			input: `
block1 "label1" "label2" {
  block2 {
    attribute = "str1"
  }
}
`,
			query: "block1.label1.label2.block2.attribute",
			want: `
block1 "label1" "label2" {
  block2 {
  }
}
`,
		},

		"AttributeInBlock4": {
			input: `
block1 "label1" "label2" {
  block2 "label3" "label4" {
  }
}
`,
			query: "block1.label1.label2.block2.label3.label4.attribute",
			want: `
block1 "label1" "label2" {
  block2 "label3" "label4" {
  }
}
`,
		},

		"AttributeInObject1": {
			input: `
object = {
  attribute = "str1"
}
`,
			query: "object.attribute",
			want: `
object = {
}
`,
		},

		"AttributeInObject2": {
			input: `
object1 = {
  object2 = {
    attribute = "str"
  }
}
`,
			query: "object1.object2.attribute",
			want: `
object1 = {
  object2 = {
  }
}
`,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			editor, err := hcledit.Read(strings.NewReader(tc.input), "")
			if err != nil {
				t.Fatal(err)
			}
			if err := editor.Delete(tc.query); err != nil {
				t.Fatal(err)
			}

			diff := cmp.Diff(tc.want, string(editor.Bytes()),
				cmpopts.AcyclicTransformer("multiline", func(s string) []string {
					return strings.Split(s, "\n")
				}),
			)

			if diff != "" {
				t.Errorf("Delete() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

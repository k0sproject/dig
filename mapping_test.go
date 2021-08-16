package dig

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestDig(t *testing.T) {
	m := Mapping{
		"foo": Mapping{
			"bar": "foobar",
		},
	}

	assert.Equal(t, "foobar", m.Dig("foo", "bar"))

	assert.Nil(t, m.Dig("foo", "non-existing", "key"))
}

func TestDigString(t *testing.T) {
	m := Mapping{
		"foo": Mapping{
			"bar": "foobar",
		},
	}

	assert.Equal(t, "foobar", m.DigString("foo", "bar"))
	assert.Equal(t, "", m.DigString("foo", "nonexisting"))
	assert.Equal(t, "", m.DigString("nonexisting", "nonexisting"))
}

func TestDigMapping(t *testing.T) {
	m := Mapping{
		"foo": Mapping{
			"bar": "foobar",
		},
	}

	assert.Equal(t, "foobar", m.DigMapping("foo")["bar"])

	m.DigMapping("foo", "baz")["dog"] = 1
	assert.Equal(t, 1, m.Dig("foo", "baz", "dog"))
	// Make sure foo.bar was left intact
	assert.Equal(t, "foobar", m.Dig("foo", "bar"))

	// Overwrite foo.bar with a new mapping
	m.DigMapping("foo", "bar")["baz"] = "hello"
	assert.Equal(t, "hello", m.Dig("foo", "bar", "baz"))
}

func TestDup(t *testing.T) {
	m := Mapping{
		"foo": Mapping{
			"bar": "foobar",
		},
		"array": []string{
			"hello",
		},
		"mappingarray": []Mapping{
			{"bar": "foobar"},
			{"foo": "barfoo"},
		},
	}

	dup := m.Dup()

	m.DigMapping("foo")["bar"] = "barbar"
	arr := m.Dig("array").([]string)
	arr = append(arr, "world")
	m["array"] = arr

	ma := m["mappingarray"].([]Mapping)
	maa := ma[0]
	maa["bar"] = "barbar"

	assert.Equal(t, "barbar", m.Dig("foo", "bar"))
	assert.Equal(t, "foobar", dup.Dig("foo", "bar"))

	a := m.Dig("array").([]string)
	b := dup.Dig("array").([]string)

	assert.Len(t, a, 2)
	assert.Len(t, b, 1)

	am := m.Dig("mappingarray").([]Mapping)
	bm := dup.Dig("mappingarray").([]Mapping)

	assert.Equal(t, "barbar", am[0]["bar"])
	assert.Equal(t, "foobar", bm[0]["bar"])
}

func TestUnmarshalYamlWithNil(t *testing.T) {
	data := `foo: null`
	var m Mapping
	err := yaml.Unmarshal([]byte(data), &m)
	assert.NoError(t, err)
	assert.Nil(t, m.Dig("foo"))
}

func ExampleMapping_Dig() {
	h := Mapping{
		"greeting": Mapping{
			"target": "world",
		},
	}
	fmt.Println("Hello,", h.Dig("greeting", "target"))
	// Output: Hello, world
}

func ExampleMapping_DigMapping() {
	h := Mapping{}
	h.DigMapping("greeting")["target"] = "world"
	fmt.Println("Hello,", h.Dig("greeting", "target"))
	// Output: Hello, world
}

func ExampleMapping_DigString() {
	h := Mapping{}
	h.DigMapping("greeting")["target"] = "world"
	fmt.Println("Hello,", h.DigString("greeting", "target"), "!")
	fmt.Println("Hello,", h.Dig("greeting", "non-existing"), "!")
	fmt.Println("Hello,", h.DigString("greeting", "non-existing"), "!")
	// Output:
	// Hello, world !
	// Hello, <nil> !
	// Hello,  !
}

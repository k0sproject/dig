package dig_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-yaml/yaml"
	"github.com/k0sproject/dig"
)

func mustBeNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func mustEqualString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func mustEqualFloat32(t *testing.T, expected, actual float32) {
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func mustBeNil(t *testing.T, actual any) {
	if actual != nil {
		t.Errorf("Expected nil, got %v", actual)
	}
}

func mustEqual(t *testing.T, expected, actual any) {
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestDig(t *testing.T) {
	m := dig.Mapping{
		"foo": dig.Mapping{
			"bar": "foobar",
		},
	}

	t.Run("fetch nested value", func(t *testing.T) {
		mustEqualString(t, "foobar", m.Dig("foo", "bar").(string))
	})

	t.Run("non-existing key should return nil", func(t *testing.T) {
		mustBeNil(t, m.Dig("foo", "non-existing"))
	})
}

func TestDigString(t *testing.T) {
	m := dig.Mapping{
		"foo": dig.Mapping{
			"bar": "foobar",
		},
	}

	t.Run("fetch nested string", func(t *testing.T) {
		mustEqualString(t, "foobar", m.DigString("foo", "bar"))
	})

	t.Run("non-existing key should return an empty string", func(t *testing.T) {
		mustEqualString(t, "", m.DigString("foo", "non-existing"))
		mustEqualString(t, "", m.DigString("non-existing", "non-existing"))
	})
}

func TestDigMapping(t *testing.T) {
	m := dig.Mapping{
		"foo": dig.Mapping{
			"bar": "foobar",
		},
	}

	t.Run("fetch nested mapping", func(t *testing.T) {
		mustEqualString(t, "foobar", m.DigMapping("foo")["bar"].(string))
	})

	t.Run("set a nested value", func(t *testing.T) {
		m.DigMapping("foo", "baz")["dog"] = 1
		mustEqual(t, 1, m.Dig("foo", "baz", "dog"))

		// Make sure foo.bar was left intact
		mustEqualString(t, "foobar", m.DigString("foo", "bar"))
	})

	t.Run("overwrite mapping", func(t *testing.T) {
		m.DigMapping("foo", "bar")["baz"] = "hello"
		mustEqualString(t, "hello", m.DigString("foo", "bar", "baz"))
		mustBeNil(t, m.Dig("foo", "bar", "dog"))
	})
}

func TestDup(t *testing.T) {
	m := dig.Mapping{
		"foo": dig.Mapping{
			"bar": "foobar",
		},
		"array": []string{
			"hello",
		},
		"mappingarray": []dig.Mapping{
			{"bar": "foobar"},
			{"foo": "barfoo"},
		},
	}

	dup := m.Dup()

	t.Run("modifying clone's values should not modify original", func(t *testing.T) {
		m.DigMapping("foo")["bar"] = "barbar"
		mustEqualString(t, "foobar", dup.DigString("foo", "bar"))
		mustEqualString(t, "barbar", m.DigString("foo", "bar"))
	})

	t.Run("modifying a cloned slice should not modify original", func(t *testing.T) {
		arr := m.Dig("array").([]string)
		arr = append(arr, "world")
		m["array"] = arr

		ma := m["mappingarray"].([]dig.Mapping)
		maa := ma[0]
		maa["bar"] = "barbar"

		mustEqual(t, 1, len(dup.Dig("array").([]string)))
		mustEqual(t, 2, len(m.Dig("array").([]string)))

		am := m.Dig("mappingarray").([]dig.Mapping)
		bm := dup.Dig("mappingarray").([]dig.Mapping)
		mustEqualString(t, "barbar", am[0]["bar"].(string))
		mustEqualString(t, "foobar", bm[0]["bar"].(string))
	})
}

func TestUnmarshalYamlWithNil(t *testing.T) {
	data := []byte(`{"foo": null}`)
	var m dig.Mapping
	mustBeNil(t, json.Unmarshal(data, &m))
	mustBeNil(t, m.Dig("foo"))
}

func TestUnmarshalYamlFloat(t *testing.T) {
	var m dig.Mapping

	err := yaml.Unmarshal([]byte(
		`
            float_32: 0.22
        `), &m)
	mustBeNoError(t, err)

	mustEqualFloat32(t, 0.22, m.Dig("float_32").(float32))
}

func ExampleMapping_Dig() {
	h := dig.Mapping{
		"greeting": dig.Mapping{
			"target": "world",
		},
	}
	fmt.Println("Hello,", h.Dig("greeting", "target"))
	// Output: Hello, world
}

func ExampleMapping_DigMapping() {
	h := dig.Mapping{}
	h.DigMapping("greeting")["target"] = "world"
	fmt.Println("Hello,", h.Dig("greeting", "target"))
	// Output: Hello, world
}

func ExampleMapping_DigString() {
	h := dig.Mapping{}
	h.DigMapping("greeting")["target"] = "world"
	fmt.Println("Hello,", h.DigString("greeting", "target"), "!")
	fmt.Println("Hello,", h.Dig("greeting", "non-existing"), "!")
	fmt.Println("Hello,", h.DigString("greeting", "non-existing"), "!")
	// Output:
	// Hello, world !
	// Hello, <nil> !
	// Hello,  !
}

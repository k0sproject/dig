package dig_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/k0sproject/dig"
)

func mustEqualString(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func mustBeNil(t *testing.T, actual any) {
	t.Helper()
	if actual != nil {
		t.Errorf("Expected nil, got %v", actual)
	}
}

func mustEqual(t *testing.T, expected, actual any) {
	t.Helper()
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

	t.Run("int value", func(t *testing.T) {
		m.DigMapping("foo")["int"] = 1
		mustEqual(t, 1, m.Dig("foo", "int"))
	})

	t.Run("float value", func(t *testing.T) {
		m.DigMapping("foo")["float"] = 0.5
		mustEqual(t, 0.5, m.Dig("foo", "float"))
	})

	t.Run("bool value", func(t *testing.T) {
		m.DigMapping("foo")["bool"] = true
		mustEqual(t, true, m.Dig("foo", "bool"))
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

func TestUnmarshalJSONWithNil(t *testing.T) {
	data := []byte(`{"foo": null}`)
	var m dig.Mapping
	mustBeNil(t, json.Unmarshal(data, &m))
	mustBeNil(t, m.Dig("foo"))
}

func TestUnmarshalJSONWithFloat(t *testing.T) {
	data := []byte(`{"foo": 0.5}`)
	var m dig.Mapping
	mustBeNil(t, json.Unmarshal(data, &m))
	val, ok := m.Dig("foo").(float64)
	mustEqual(t, true, ok)
	mustEqual(t, 0.5, val)
}

func TestUnmarshalJSONWithSliceOfMaps(t *testing.T) {
	data := []byte(`{"foo": [{"bar": "baz"}]}`)
	var m dig.Mapping
	mustBeNil(t, json.Unmarshal(data, &m))
	val, ok := m.Dig("foo").([]any)
	mustEqual(t, true, ok)
	obj, ok := val[0].(dig.Mapping)
	mustEqual(t, true, ok)
	mustEqual(t, "baz", obj["bar"])
}

func TestMerge(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		m := dig.Mapping{
			"foo": "bar",
			"bar": "baz",
			"nested": dig.Mapping{
				"foo": "bar",
			},
		}
		other := dig.Mapping{
			"foo": "baz",
			"bar": nil,
			"nested": dig.Mapping{
				"foo": "baz",
				"bar": "foo",
			},
		}
		m.Merge(other)
		mustEqualString(t, "bar", m.DigString("foo"))
		mustEqualString(t, "bar", m.DigString("nested", "foo"))
		mustEqualString(t, "foo", m.DigString("nested", "bar"))
		mustEqualString(t, "baz", m.DigString("bar"))
	})
	t.Run("overwrite", func(t *testing.T) {
		m := dig.Mapping{
			"foo": "bar",
			"bar": "baz",
			"nested": dig.Mapping{
				"foo": "bar",
			},
		}
		other := dig.Mapping{
			"foo": "baz",
			"bar": nil,
			"nested": dig.Mapping{
				"foo": "baz",
				"bar": "foo",
			},
		}
		m.Merge(other, dig.WithOverwrite())
		mustEqualString(t, "baz", m.DigString("foo"))
		mustEqualString(t, "baz", m.DigString("bar"))
		mustEqualString(t, "baz", m.DigString("nested", "foo"))
		mustEqualString(t, "foo", m.DigString("nested", "bar"))
	})
	t.Run("nillify", func(t *testing.T) {
		m := dig.Mapping{
			"foo": "bar",
			"bar": "baz",
			"nested": dig.Mapping{
				"foo": "bar",
			},
		}
		other := dig.Mapping{
			"foo": "baz",
			"bar": nil,
			"nested": dig.Mapping{
				"foo": nil,
				"bar": "foo",
			},
		}
		m.Merge(other, dig.WithNillify())
		mustEqualString(t, "bar", m.DigString("foo"))
		mustBeNil(t, m.Dig("bar"))
		mustBeNil(t, m.Dig("nested", "foo"))
		mustEqualString(t, "foo", m.DigString("nested", "bar"))
	})
	t.Run("overwrite+nillify", func(t *testing.T) {
		m := dig.Mapping{
			"foo": "bar",
			"bar": "baz",
			"nested": dig.Mapping{
				"foo": "bar",
			},
		}
		other := dig.Mapping{
			"foo": "baz",
			"bar": nil,
			"nested": dig.Mapping{
				"foo": nil,
				"bar": "foo",
			},
		}
		m.Merge(other, dig.WithOverwrite(), dig.WithNillify())
		mustEqualString(t, "baz", m.DigString("foo"))
		mustBeNil(t, m.Dig("bar"))
		mustBeNil(t, m.Dig("nested", "foo"))
		mustEqualString(t, "foo", m.DigString("nested", "bar"))
	})
	t.Run("slice append", func(t *testing.T) {
		m := dig.Mapping{
			"foo": []string{"bar"},
			"bar": []string{"baz"},
			"map": []dig.Mapping{
				{"foo": "bar"},
			},
		}
		other := dig.Mapping{
			"foo": []string{"baz"},
			"bar": nil,
			"map": []dig.Mapping{
				{"foo": "baz"},
			},
		}
		m.Merge(other, dig.WithAppend())
		foo := m.Dig("foo").([]string)
		mustEqual(t, 2, len(foo))
		mustEqual(t, "bar", foo[0])
		mustEqual(t, "baz", foo[1])
		mustEqual(t, "baz", m.Dig("bar").([]string)[0])
		mustEqual(t, 1, len(m.Dig("bar").([]string)))
		mapping := m.Dig("map").([]dig.Mapping)
		mustEqual(t, 2, len(mapping))
		mustEqual(t, "bar", mapping[0]["foo"])
		mustEqual(t, "baz", mapping[1]["foo"])
	})
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

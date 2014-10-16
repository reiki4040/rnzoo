package goltsv

import (
	"testing"
)

type ExceptParseLtsv struct {
	Ltsv   string
	Expect map[string]string
}
type ExceptParseLtsvFilter struct {
	Ltsv   string
	Filter []string
	Expect map[string]string
}

var (
	ForParseLtsv = []ExceptParseLtsv{
		ExceptParseLtsv{
			Ltsv:   "",
			Expect: map[string]string{},
		},
		ExceptParseLtsv{
			Ltsv: "key1:value1",
			Expect: map[string]string{
				"key1": "value1",
			},
		},
		ExceptParseLtsv{
			Ltsv: "key1:",
			Expect: map[string]string{
				"key1": "",
			},
		},
		ExceptParseLtsv{
			Ltsv: ":value1",
			Expect: map[string]string{
				"": "value1",
			},
		},
		ExceptParseLtsv{
			Ltsv: ":",
			Expect: map[string]string{
				"": "",
			},
		},
		ExceptParseLtsv{
			Ltsv: "key1:value1\tkey2:value2",
			Expect: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	ForParseLtsvFilter = []ExceptParseLtsvFilter{
		ExceptParseLtsvFilter{
			Ltsv:   "",
			Filter: []string{},
			Expect: map[string]string{},
		},
		ExceptParseLtsvFilter{
			Ltsv:   "",
			Filter: []string{"key"},
			Expect: map[string]string{},
		},
		ExceptParseLtsvFilter{
			Ltsv:   "key1:value1",
			Filter: []string{},
			Expect: map[string]string{},
		},
		ExceptParseLtsvFilter{
			Ltsv:   "key1:value1",
			Filter: []string{"key1"},
			Expect: map[string]string{
				"key1": "value1",
			},
		},
		ExceptParseLtsvFilter{
			Ltsv:   "key1:value1",
			Filter: []string{"key2"},
			Expect: map[string]string{},
		},
		ExceptParseLtsvFilter{
			Ltsv:   "key1:value1\tkey2:value2\tkey3:value3",
			Filter: []string{},
			Expect: map[string]string{},
		},
		ExceptParseLtsvFilter{
			Ltsv:   "key1:value1\tkey2:value2\tkey3:value3",
			Filter: []string{"key3"},
			Expect: map[string]string{
				"key3": "value3",
			},
		},
		ExceptParseLtsvFilter{
			Ltsv:   "key1:value1\tkey2:value2\tkey3:value3",
			Filter: []string{"key1", "key3"},
			Expect: map[string]string{
				"key1": "value1",
				"key3": "value3",
			},
		},
		ExceptParseLtsvFilter{
			Ltsv:   "key1:value1\tkey2:value2\tkey3:value3",
			Filter: []string{"key1", "key2", "key3"},
			Expect: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}
)

func TestParseLtsv(t *testing.T) {
	for _, input := range ForParseLtsv {

		result := ParseLtsv(input.Ltsv)

		if len(result) != len(input.Expect) {
			t.Errorf("failed: expect %d but %d\n", len(input.Expect), len(result))
			t.Errorf("Expect:\n%v\nResult:\n%v\n", input.Expect, result)
		}

		for k, v := range input.Expect {
			rv, exists := result[k]
			if !exists {
				t.Errorf("not has key %s in result.\n", k)
			}
			if rv != v {
				t.Errorf("expect value %s => %s, but %s\n", k, input.Expect[k], rv)
			}
		}
	}
}

func TestArray2set(t *testing.T) {
	array := []string{"key1", "key2"}
	slice := array2set(array)

	if len(slice) != len(array) {
		t.Errorf("mismatch length array %d, set %d\n", len(array), len(slice))
	}

	for _, key := range array {
		if _, ok := slice[key]; !ok {
			t.Errorf("not exists in set %s\n", key)
		}
	}
}

func TestParseLtsvFilter(t *testing.T) {
	for _, input := range ForParseLtsvFilter {

		result := ParseLtsvFilter(input.Ltsv, input.Filter...)

		if len(result) != len(input.Expect) {
			t.Errorf("failed: expect %d but %d\n", len(input.Expect), len(result))
			t.Errorf("Expect:\n%v\nResult:\n%v\n", input.Expect, result)
		}

		for k, v := range input.Expect {
			rv, exists := result[k]
			if !exists {
				t.Errorf("not has key %s in result.\n", k)
			}
			if rv != v {
				t.Errorf("expect value %s => %s, but %s\n", k, input.Expect[k], rv)
			}
		}
	}
}

func TestMap2Ltsv(t *testing.T) {
	for _, input := range ForParseLtsv {

		ltsv := Map2Ltsv(input.Expect)

		result := ParseLtsv(ltsv)
		if len(result) != len(input.Expect) {
			t.Errorf("failed: expect %d but %d\n", len(input.Expect), len(result))
			t.Errorf("Expect:\n%v\nResult:\n%v\n", input.Expect, result)
		}

		for k, v := range input.Expect {
			rv, exists := result[k]
			if !exists {
				t.Errorf("not has key %s in result.\n", k)
			}
			if rv != v {
				t.Errorf("expect value %s => %s, but %s\n", k, input.Expect[k], rv)
			}
		}
	}
}

type ExceptMap2OrderedLtsv struct {
	Items  map[string]string
	Filter []string
	Expect string
}

var (
	ForMap2OrderdLtsv = []ExceptMap2OrderedLtsv{
		ExceptMap2OrderedLtsv{
			Items:  nil,
			Filter: []string{},
			Expect: "",
		},
		ExceptMap2OrderedLtsv{
			Items:  map[string]string{},
			Filter: []string{},
			Expect: "",
		},
		ExceptMap2OrderedLtsv{
			Items:  map[string]string{},
			Filter: []string{"key1"},
			Expect: "",
		},
		ExceptMap2OrderedLtsv{
			Items: map[string]string{
				"key1": "value1",
			},
			Filter: []string{},
			Expect: "",
		},
		ExceptMap2OrderedLtsv{
			Items: map[string]string{
				"key1": "value1",
			},
			Filter: []string{"key1"},
			Expect: "key1:value1",
		},
		ExceptMap2OrderedLtsv{
			Items: map[string]string{
				"key1": "value1",
			},
			Filter: []string{"key1", "key1"},
			Expect: "key1:value1\tkey1:value1",
		},
		ExceptMap2OrderedLtsv{
			Items: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			Filter: []string{"key3", "key1"},
			Expect: "key3:value3\tkey1:value1",
		},
		ExceptMap2OrderedLtsv{
			Items: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			Filter: []string{"key2", "key1", "key3"},
			Expect: "key2:value2\tkey1:value1\tkey3:value3",
		},
	}
)

func TestMap2OrderedLtsv(t *testing.T) {
	for _, input := range ForMap2OrderdLtsv {

		ltsv := Map2OrderedLtsv(input.Items, input.Filter...)

		if ltsv != input.Expect {
			t.Errorf("result:%s\nExpect:%s\n", ltsv, input.Expect)
		}
	}
}

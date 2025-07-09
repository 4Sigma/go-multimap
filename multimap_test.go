package multimap

import (
	"encoding/json"
	"reflect"
	"testing"
)

func intEquals(a, b int) bool {
	return a == b
}

func TestAddAndGet(t *testing.T) {
	m := New[string, int](intEquals)
	m.Add("a", 1)
	m.Add("a", 2)
	m.Add("a", 1) // duplicate, should not be added

	vals := m.Get("a")
	expected := []int{1, 2}
	if !reflect.DeepEqual(vals, expected) {
		t.Errorf("expected %v, got %v", expected, vals)
	}
}

func TestRemove(t *testing.T) {
	m := New[string, int](intEquals)
	m.Add("a", 1)
	m.Add("b", 2)
	m.Remove("a")
	if m.HasKey("a") {
		t.Errorf("expected key 'a' to be removed")
	}
	if !m.HasKey("b") {
		t.Errorf("expected key 'b' to exist")
	}
}

func TestRemoveValue(t *testing.T) {
	m := New[string, int](intEquals)
	m.Add("a", 1)
	m.Add("a", 2)
	m.RemoveValue("a", 1)
	vals := m.Get("a")
	expected := []int{2}
	if !reflect.DeepEqual(vals, expected) {
		t.Errorf("expected %v, got %v", expected, vals)
	}
	m.RemoveValue("a", 2)
	if m.HasKey("a") {
		t.Errorf("expected key 'a' to be removed after last value removed")
	}
}

func TestHasAndHasKey(t *testing.T) {
	m := New[string, int](intEquals)
	m.Add("a", 1)
	if !m.Has("a", 1) {
		t.Errorf("expected to have value 1 for key 'a'")
	}
	if m.Has("a", 2) {
		t.Errorf("did not expect to have value 2 for key 'a'")
	}
	if !m.HasKey("a") {
		t.Errorf("expected key 'a' to exist")
	}
	if m.HasKey("b") {
		t.Errorf("did not expect key 'b' to exist")
	}
}

func TestKeys(t *testing.T) {
	m := New[string, int](intEquals)
	m.Add("a", 1)
	m.Add("b", 2)
	keys := m.Keys()
	expected := []string{"a", "b"}
	// Order is not guaranteed
	keyMap := map[string]bool{}
	for _, k := range keys {
		keyMap[k] = true
	}
	for _, k := range expected {
		if !keyMap[k] {
			t.Errorf("expected key %s in keys", k)
		}
	}
}

func TestLenAndCount(t *testing.T) {
	m := New[string, int](intEquals)
	if m.Len() != 0 || m.Count() != 0 {
		t.Errorf("expected empty map")
	}
	m.Add("a", 1)
	m.Add("a", 2)
	m.Add("b", 3)
	if m.Len() != 2 {
		t.Errorf("expected 2 keys, got %d", m.Len())
	}
	if m.Count() != 3 {
		t.Errorf("expected 3 values, got %d", m.Count())
	}
}

func TestClear(t *testing.T) {
	m := New[string, int](intEquals)
	m.Add("a", 1)
	m.Clear()
	if m.Len() != 0 {
		t.Errorf("expected map to be cleared")
	}
}

func TestForEach(t *testing.T) {
	m := New[string, int](intEquals)
	m.Add("a", 1)
	m.Add("a", 2)
	m.Add("b", 3)
	result := map[string][]int{}
	m.ForEach(func(k string, v int) {
		result[k] = append(result[k], v)
	})
	if !reflect.DeepEqual(result["a"], []int{1, 2}) || !reflect.DeepEqual(result["b"], []int{3}) {
		t.Errorf("unexpected result from ForEach: %v", result)
	}
}

func TestClone(t *testing.T) {
	m := New[string, int](intEquals)
	m.Add("a", 1)
	m.Add("b", 2)
	clone := m.Clone()
	if !m.Equal(clone) {
		t.Errorf("expected clone to be equal to original")
	}
	clone.Add("a", 3)
	if m.Equal(clone) {
		t.Errorf("expected clone to differ after modification")
	}
}

func TestEqual(t *testing.T) {
	m1 := New[string, int](intEquals)
	m2 := New[string, int](intEquals)
	if !m1.Equal(m2) {
		t.Errorf("expected two empty multimaps to be equal")
	}
	m1.Add("a", 1)
	m2.Add("a", 1)
	if !m1.Equal(m2) {
		t.Errorf("expected multimaps to be equal")
	}
	m2.Add("a", 2)
	if m1.Equal(m2) {
		t.Errorf("expected multimaps to differ")
	}
}

func TestMarshalUnmarshalJSON(t *testing.T) {
	m := New[string, int](intEquals)
	m.Add("a", 1)
	m.Add("a", 2)
	m.Add("b", 3)
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	m2 := New[string, int](intEquals)
	if err := json.Unmarshal(data, m2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !m.Equal(m2) {
		t.Errorf("expected unmarshaled multimap to equal original")
	}
}

func TestNewFromJSON(t *testing.T) {
	raw := []byte(`{"a":[1,2],"b":[3]}`)
	m, err := NewFromJSON[string, int](raw, intEquals)
	if err != nil {
		t.Fatalf("NewFromJSON error: %v", err)
	}
	expected := New[string, int](intEquals)
	expected.Add("a", 1)
	expected.Add("a", 2)
	expected.Add("b", 3)
	if !m.Equal(expected) {
		t.Errorf("expected %v, got %v", expected.data, m.data)
	}
}

func TestEdgeCases(t *testing.T) {
	m := New[string, int](intEquals)
	// Remove non-existent key
	m.Remove("nope")
	// RemoveValue on non-existent key
	m.RemoveValue("nope", 1)
	// Get on non-existent key
	vals := m.Get("nope")
	if len(vals) != 0 {
		t.Errorf("expected empty slice, got %v", vals)
	}
	// Has on non-existent key
	if m.Has("nope", 1) {
		t.Errorf("expected false for Has on non-existent key")
	}
}

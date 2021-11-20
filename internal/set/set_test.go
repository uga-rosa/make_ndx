package set

import (
	"testing"
)

func TestSetMethod(t *testing.T) {
	set := New("Alice", "Bob")
	newExpected := "{Alice, Bob}"
	if set.String() != newExpected {
		t.Fatalf("New method won't work. got=%q, want=%q", set.String(), newExpected)
	}

	set.Add("Alice")
	set.Add("Carol")
	addExpected := "{Alice, Bob, Carol}"
	if set.String() != addExpected {
		t.Fatalf("Add method won't work. got=%q, want=%q", set.String(), addExpected)
	}

	set.Remove("Alice")
	set.Remove("Dave")
	removeExpected := "{Bob, Carol}"
	if set.String() != removeExpected {
		t.Fatalf("Remove method won't work. got=%q, want=%q", set.String(), removeExpected)
	}

	if set.Contains("Alice") {
		t.Fatalf("Contains method won't work. 'Alice' doesn't contained.")
	}
}

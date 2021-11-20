package set

import (
	"bytes"
	"strings"

	"github.com/mattn/natural"
)

type Set map[string]struct{}

func New(str ...string) *Set {
	set := make(Set)
	for _, s := range str {
		set.Add(s)
	}
	return &set
}

func (s *Set) Add(e string) {
	(*s)[e] = struct{}{}
}

func (s *Set) Remove(e string) {
	delete(*s, e)
}

func (s *Set) Contains(e string) bool {
	if _, ok := (*s)[e]; ok {
		return true
	}
	return false
}

func (s *Set) ToSlice() []string {
	slice := make([]string, 0, len(*s))
	for k := range *s {
		slice = append(slice, k)
	}
	natural.Sort(slice)
	return slice
}

func (s *Set) String() string {
	var out bytes.Buffer

	out.WriteString("{")
	out.WriteString(strings.Join(s.ToSlice(), ", "))
	out.WriteString("}")

	return out.String()
}

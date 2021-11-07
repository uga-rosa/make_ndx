package cmd

import (
	"github.com/mattn/natural"
)

type Set map[string]interface{}

func (s *Set) Contains(e string) bool {
	if _, ok := (*s)[e]; ok {
		return true
	}
	return false
}

func (s *Set) Remove(e string) {
	if _, ok := (*s)[e]; ok {
		delete(*s, e)
	}
}

func (s *Set) Add(e string) {
	(*s)[e] = struct{}{}
}

func (s *Set) ToSlice() []string {
	slice := make([]string, 0, len(*s))
	for k := range *s {
		slice = append(slice, k)
	}
	natural.Sort(slice)
	return slice
}

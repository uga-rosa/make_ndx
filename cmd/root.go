/*
Copyright © 2021 uga-rosa uga6603@gmail.com
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mattn/natural"
	"github.com/spf13/cobra"
)

type Options struct {
	f string
	o string
	k string
}

var option = &Options{}

var rootCmd = &cobra.Command{
	Use:   "make_ndx",
	Short: "Finer gmx make_ndx",
	Run: func(cmd *cobra.Command, args []string) {
		makeNdx(option.f, option.o, option.k)
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringVarP(&option.f, "input", "f", "", "Structure file: gro, pdb (required)")
	rootCmd.MarkFlagRequired("f")
	rootCmd.Flags().StringVarP(&option.o, "output", "o", "index.ndx", "Index file")
	rootCmd.Flags().StringVarP(&option.k, "combine", "c", "resnum", "combine rule. 'resnum'|'atomname'")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func makeNdx(input, output, key string) {
	if key != "resnum" && key != "atomname" {
		panic(errors.New("Invalid option k. k must be 'resnum' or 'atomname'."))
	}
	atoms := getAtoms(input)
	resNames := getAllResName(atoms)
	choicedResName := selectName(resNames, "residue name")
	choiced := make(Choiced)
	for k := range choicedResName {
		choiced[k] = &breakdown{
			atomnames: selectName(getAtomName(atoms, k), "atom name"),
			resnums:   selectName(getResNum(atoms, k), "residue number"),
		}
	}
	group := combine(choiced, atoms, key)
	writeGroups(group, output)
}

type (
	Atom struct {
		resNum   string
		atomName string
		atomNum  string
	}
	Atoms map[string][]Atom // key is resName

	breakdown struct {
		atomnames Set
		resnums   Set
	}
	Choiced map[string]*breakdown // key is resName
	Group   map[string][]string   // key is group name, value is slice of atom numbers
)

func getAtoms(input string) Atoms {
	ext := filepath.Ext(input)
	if ext != ".gro" {
		panic(fmt.Sprint("This format is not supported: ", ext))
	}

	lines := readlines(input)
	atoms := make(Atoms)
	for i := 2; i < len(lines)-1; i++ {
		resName := strings.TrimSpace(lines[i][5:10])
		if _, ok := atoms[resName]; !ok {
			atoms[resName] = make([]Atom, 0)
		}
		atoms[resName] = append(atoms[resName], Atom{
			resNum:   strings.TrimSpace(lines[i][0:5]),
			atomName: strings.TrimSpace(lines[i][10:15]),
			atomNum:  strings.TrimSpace(lines[i][15:20]),
		})
	}

	return atoms
}

func readlines(filename string) []string {
	file, err := os.Open(filename)
	check(err)
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func getAllResName(atoms Atoms) Set {
	resNames := make(Set)
	for resName := range atoms {
		resNames.Add(resName)
	}
	return resNames
}

func getResNum(atoms Atoms, resName string) Set {
	resNums := make(Set)
	for _, atom := range atoms[resName] {
		resNums.Add(atom.resNum)
	}
	return resNums
}

func getAtomName(atoms Atoms, resName string) Set {
	atomNames := make(Set)
	for _, atom := range atoms[resName] {
		atomNames.Add(atom.atomName)
	}
	return atomNames
}

func selectName(candidates Set, kind string) Set {
	fmt.Println(candidates.ToSlice())
	choiced := make(Set)
	add := true
	for {
		fmt.Print("Select a ", kind, ". > ")
		if add {
			selectString(&candidates, &choiced)
		} else {
			selectString(&choiced, &candidates)
		}
		fmt.Print("Choiced: ")
		display(&choiced)
		next := addOrRemoveOrNo()
		if next == "Add" {
			add = true
		} else if next == "Remove" {
			add = false
		} else {
			break
		}
	}
	return choiced
}

func display(s *Set) {
	fmt.Println(strings.Join(s.ToSlice(), ", "))
}

func selectString(candidates, choiced *Set) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	in := scanner.Text()
	pat := regexp.MustCompile(`(\d+),(\d+)`)
	if candidates.Contains(in) {
		choiced.Add(in)
	} else if numbers := pat.FindStringSubmatch(in); len(numbers) == 3 {
		start, _ := strconv.Atoi(numbers[1])
		end, _ := strconv.Atoi(numbers[2])
		for i := start; i <= end; i++ {
			choice := pat.ReplaceAllString(in, strconv.Itoa(i))
			if candidates.Contains(choice) {
				candidates.Remove(choice)
				choiced.Add(choice)
			}
		}
	} else {
		pat = regexp.MustCompile(in)
		for c := range *candidates {
			if pat.MatchString(c) {
				candidates.Remove(c)
				choiced.Add(c)
			}
		}
	}
}

func addOrRemoveOrNo() string {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Continue? [add/remove/no] > ")
		scanner.Scan()
		switch scanner.Text() {
		case "Add", "add", "A", "a":
			return "Add"
		case "Remove", "remove", "R", "r":
			return "Remove"
		case "No", "no", "N", "n":
			return "No"
		default:
			fmt.Println("Invalid input.")
		}
	}
}

func combine(c Choiced, atoms Atoms, key string) Group {
	atomNums := make(map[string][]string)
	switch key {
	case "resnum":
		for resName, bd := range c {
			for resNum := range bd.resnums {
				groupName := resName + resNum
				group := make([]string, 0)
				for _, atom := range atoms[resName] {
					if atom.resNum == resNum && bd.atomnames.Contains(atom.atomName) {
						group = append(group, atom.atomNum)
					}
				}
				atomNums[groupName] = group
			}
		}
	case "atomname":
		for resName, bd := range c {
			for atomName := range bd.atomnames {
				groupName := resName + atomName
				group := make([]string, 0)
				for _, atom := range atoms[resName] {
					if atom.atomName == atomName && bd.resnums.Contains(atom.resNum) {
						group = append(group, atom.atomNum)
					}
				}
				atomNums[groupName] = group
			}
		}
	}
	return atomNums
}

func writeGroups(g Group, output string) {
	if fileExists(output) {
		counter := 1
		backup := fmt.Sprint("#", output, ".", counter, "#")
		for fileExists(backup) {
			counter += 1
			backup = fmt.Sprint("#", output, ".", counter, "#")
		}
		os.Rename(output, backup)
		fmt.Println("Back Off! I just backed up", output, "to", backup)
	}
	file, err := os.Create(output)
	check(err)
	defer file.Close()
	for _, name := range g.GroupKeyToSlice() {
		_, err := file.WriteString("[ " + name + " ]\n")
		check(err)
		numbers := g[name]
		for i := 0; i < len(numbers); i += 10 {
			ends := i + 10
			if ends >= len(numbers) {
				ends = len(numbers) - 1
			}
			_, err := file.WriteString(strings.Join(numbers[i:ends], " ") + "\n")
			check(err)
		}
		file.WriteString("\n")
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (g *Group) GroupKeyToSlice() []string {
	slice := make([]string, 0, len(*g))
	for k := range *g {
		slice = append(slice, k)
	}
	natural.Sort(slice)
	return slice
}

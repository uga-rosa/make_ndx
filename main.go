package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mattn/natural"
	"github.com/uga-rosa/make_ndx/internal/set"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "make_ndx",
		Usage: "Finer gmx make_ndx",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"f"},
				Value:    "",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "index.ndx",
			},
			&cli.StringFlag{
				Name:    "combine",
				Aliases: []string{"c"},
				Value:   "resnum",
			},
		},
		Action: makeNdx,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

type (
	Atom struct {
		resNum   string
		atomName string
		atomNum  string
	}
	Atoms map[string][]Atom // key is resName

	selectedSet struct {
		resNums   set.Set
		atomNames set.Set
	}
	Choiced map[string]selectedSet

	Group map[string][]string // key is group name, value is slice of atom numbers.
)

func makeNdx(c *cli.Context) error {
	input := c.String("input")
	output := c.String("output")
	combine := c.String("combine")
	if err := argCheck(input, output, combine); err != nil {
		return err
	}

	lines, err := readlines(input)
	if err != nil {
		return err
	}
	atoms := getAtoms(lines)
	resNames := atoms.getAllResName()
	choicedResNames := selectFromSet(resNames, "residue name")
	choiced := make(Choiced)
	for k := range *choicedResNames {
		choiced[k] = selectedSet{
			*selectFromSet(atoms.getResNum(k), "residue number"),
			*selectFromSet(atoms.getAtomName(k), "atom name"),
		}
	}
	group := atoms.combine(choiced, combine)
	err = writeGroup(group, output)
	if err != nil {
		return err
	}

	return nil
}

func argCheck(input, output, combine string) error {
	ext := filepath.Ext(input)
	if ext != ".gro" {
		return fmt.Errorf("Argument error (input): %s is invalid file extension", ext)
	}

	ext = filepath.Ext(output)
	if ext != ".ndx" {
		return fmt.Errorf("Argument error (output): %s is invalid file extension", ext)
	}

	if !set.New("resname", "atomname").Contains(combine) {
		return fmt.Errorf("Argument error (combine): %s is not `resname` or `atomname`", ext)
	}

	return nil
}

func getAtoms(lines []string) Atoms {
	atoms := make(Atoms)
	for i := 2; i < len(lines)-1; i++ {
		resName := strings.TrimSpace(lines[i][5:10])
		if _, ok := atoms[resName]; !ok {
			atoms[resName] = make([]Atom, 0)
		}
		atoms[resName] = append(atoms[resName], Atom{
			resNum:   strings.TrimSpace(lines[i][0:5]),
			atomName: strings.TrimSpace(lines[i][10:15]),
			atomNum:  lines[i][15:20],
		})
	}

	return atoms
}

func readlines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func (atoms *Atoms) getAllResName() *set.Set {
	resNames := set.New()
	for resName := range *atoms {
		resNames.Add(resName)
	}
	return resNames
}

func (atoms *Atoms) getResNum(resName string) *set.Set {
	resNums := set.New()
	for _, atom := range (*atoms)[resName] {
		resNums.Add(atom.resNum)
	}
	return resNums
}

func (atoms *Atoms) getAtomName(resName string) *set.Set {
	atomNames := set.New()
	for _, atom := range (*atoms)[resName] {
		atomNames.Add(atom.atomName)
	}
	return atomNames
}

func selectFromSet(candidates *set.Set, kind string) *set.Set {
	fmt.Println(candidates.String())
	choiced := set.New()
	add := true
	for {
		fmt.Print("Select a ", kind, ". > ")
		if add {
			selectString(candidates, choiced)
		} else {
			selectString(choiced, candidates)
		}
		fmt.Print("Choiced: ", choiced.String())
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

func selectString(candidates, choiced *set.Set) {
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

func (atoms *Atoms) combine(c Choiced, key string) Group {
	atomNums := make(map[string][]string)
	switch key {
	case "resnum":
		for resName, bd := range c {
			for resNum := range bd.resNums {
				groupName := resName + resNum
				group := make([]string, 0)
				for _, atom := range (*atoms)[resName] {
					if atom.resNum == resNum && bd.atomNames.Contains(atom.atomName) {
						group = append(group, atom.atomNum)
					}
				}
				atomNums[groupName] = group
			}
		}
	case "atomname":
		for resName, bd := range c {
			for atomName := range bd.atomNames {
				groupName := resName + atomName
				group := make([]string, 0)
				for _, atom := range (*atoms)[resName] {
					if atom.atomName == atomName && bd.resNums.Contains(atom.resNum) {
						group = append(group, atom.atomNum)
					}
				}
				atomNums[groupName] = group
			}
		}
	}
	return atomNums
}

func writeGroup(g Group, output string) error {
	if fileExists(output) {
		counter := 1
		backup := fmt.Sprint("#", output, ".", counter, "#")
		for fileExists(backup) {
			counter += 1
			backup = fmt.Sprint("#", output, ".", counter, "#")
		}
		err := os.Rename(output, backup)
		if err != nil {
			return err
		}
		fmt.Println("Back Off! I just backed up", output, "to", backup)
	}

	var out bytes.Buffer

	for _, name := range g.GroupKeyToSlice() {
		out.WriteString("[ " + name + " ]")
		numbers := g[name]
		for i := 0; i < len(numbers); i += 15 {
			ends := i + 15
			if ends > len(numbers) {
				ends = len(numbers)
			}
			out.WriteString(strings.Join(numbers[i:ends], " "))
		}
		out.WriteString("")
	}

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	file.WriteString(out.String())
	file.Close()

	return nil
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

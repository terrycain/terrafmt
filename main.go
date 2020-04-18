package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var cli struct {
	IndentLength int      `help:"Indent size in spaces" default:"2"`
	Recursive    bool     `help:"Search recurively for .tf files"`
	Check        bool     `help:"Check files dont require modification, returns 0 when no changes are required, 1 when changes are needed"`
	Diff         bool     `help:"Dont modify files but show diff of the changes"`
	Paths        []string `arg optional name:"path" help:"Paths or files to format" type:"path"`
}

func main() {
	kong.Parse(&cli,
		kong.Name("terrafmt"),
		kong.Description("Formats terraform files. If no path is specified, the current working directory is used."),
		kong.UsageOnError(),
	)

	cwd, err := os.Getwd()
	if len(cli.Paths) == 0 && err != nil {
		fmt.Println("Failed to get cwd", err)
		os.Exit(1)
	}

	files := FindFiles(cwd, cli.Recursive, cli.Paths)
	// fmt.Println(files)

	modifyFiles := !cli.Check && !cli.Diff

	total := len(files)
	changed := 0
	for _, file := range files {
		orignal, formatted := FormatFile(file) // TODO add options
		if orignal != formatted {
			changed++

			if modifyFiles {
				fmt.Println("Updating %v", file)
				writeFile(file, formatted)
			} else if cli.Diff {
				fmt.Println(fmt.Sprintf("File %v:", file))
				printDiff(orignal, formatted)
			}
		}
	}
	unchanged := total - changed

	if modifyFiles {
		fmt.Println(fmt.Sprintf("%v %v reformatted, %v %v left unchanged", changed, filePlural(changed), unchanged, filePlural(unchanged)))
	} else {
		fmt.Println(fmt.Sprintf("%v %v would have been reformatted, %v %v left unchanged", changed, filePlural(changed), unchanged, filePlural(unchanged)))
		if changed != 0 {
			os.Exit(1)
		}
	}
}

func filePlural(count int) string {
	if count == 1 {
		return "file"
	}
	return "files"
}

// FindFiles Finds .tf files
func FindFiles(cwd string, recursive bool, paths []string) []string {
	files := []string{}

	if len(paths) == 0 {
		paths = append(paths, cwd)
	}

	for _, currentPath := range paths {
		fi, err := os.Stat(currentPath)
		if err != nil {
			fmt.Println(fmt.Sprintf("Cannot access %v", currentPath))
			// TODO log err
			continue
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			dirFiles, err := ioutil.ReadDir(currentPath)
			if err != nil {
				fmt.Println(fmt.Sprintf("Failed to list files in %v", currentPath))
				continue
			}

			newFiles := []string{}
			for _, dirItem := range dirFiles {
				if (recursive && dirItem.IsDir()) || (!dirItem.IsDir() && strings.HasSuffix(dirItem.Name(), ".tf")) {
					newFiles = append(newFiles, path.Join(currentPath, dirItem.Name()))
				}
			}

			files = append(files, FindFiles(cwd, recursive, newFiles)...)
		case mode.IsRegular():
			if strings.HasSuffix(currentPath, ".tf") {
				files = append(files, currentPath)
			}
		default:
			fmt.Println(fmt.Sprintf("File %v is not a regular file"))
		}
	}

	return files
}

func printDiff(orig, new string) {

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(orig, new, false)

	var buff bytes.Buffer
	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			_, _ = buff.WriteString("\x1b[102m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffDelete:
			_, _ = buff.WriteString("\x1b[41m")
			_, _ = buff.WriteString(text)
			_, _ = buff.WriteString("\x1b[0m")
		case diffmatchpatch.DiffEqual:
			_, _ = buff.WriteString(text)
		}
	}

	fmt.Println(buff.String())
}

func writeFile(filepath, data string) {
	ioutil.WriteFile(filepath, []byte(data))
}

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func getDirFiles(out io.Writer, prefix, pwd string, printFiles bool) {
	files, err := ioutil.ReadDir(pwd)
	if err != nil {
		fmt.Println("Panic happend:", err)
	}
	if !printFiles {
		printOnliDir := []os.FileInfo{}
		for _, file := range files {
			if file.IsDir() {
				printOnliDir = append(printOnliDir, file)
			}
		}
		files = printOnliDir
	}
	length := len(files)
	for i, file := range files {
		if file.Name() == ".DS_Store" {
			continue
		} else if file.IsDir() {
			var prefixNew string
			if length > i+1 {
				fmt.Fprintf(out, prefix+"├───%s\n", file.Name())
				prefixNew = prefix + "│\t"
			} else {
				fmt.Fprintf(out, prefix+"└───%s\n", file.Name())
				prefixNew = prefix + "\t"
			}
			getDirFiles(out, prefixNew, filepath.Join(pwd, file.Name()), printFiles)
		} else if printFiles {
			if file.Size() > 0 {
				if length > i+1 {
					fmt.Fprintf(out, prefix+"├───%s (%vb)\n", file.Name(), file.Size())
				} else {
					fmt.Fprintf(out, prefix+"└───%s (%vb)\n", file.Name(), file.Size())
				}
			} else {
				if length > i+1 {
					fmt.Fprintf(out, prefix+"├───%s (empty)\n", file.Name())
				} else {
					fmt.Fprintf(out, prefix+"└───%s (empty)\n", file.Name())
				}
			}
		}
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	getDirFiles(out, "", path, printFiles)
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

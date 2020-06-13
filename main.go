package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const (
	elementPrefix     = "├───"
	lastElementPrefix = "└───"
	verticalPrefix    = "│"
	tab               = "\t"
	lineBreak         = "\n"
)

func buildTreeLine(fileInfo os.FileInfo, lastElement bool, prefix string) []byte {
	if lastElement {
		prefix += lastElementPrefix
	} else {
		prefix += elementPrefix
	}

	if fileInfo.IsDir() {
		return []byte(prefix + fileInfo.Name() + lineBreak)
	}

	if fileInfo.Size() == 0 {
		return []byte(prefix + fileInfo.Name() + " (empty)" + lineBreak)
	}
	return []byte(prefix + fileInfo.Name() + fmt.Sprintf(" (%db)\n", fileInfo.Size()))
}

// getFileInfos returns an *os.File instance, a list of os.FileInfo structs and error
func getFileInfos(path string) (*os.File, []os.FileInfo, error) {
	// open file descriptor
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	// get files and dirs list
	fileInfos, err := file.Readdir(0)
	if err != nil {
		return nil, nil, err
	}

	// sort list of files and dirs alphabetically
	sort.Slice(fileInfos, func(i int, j int) bool {
		return fileInfos[i].Name() < fileInfos[j].Name()
	})
	return file, fileInfos, nil
}

// getLastElementIndex evaluates index of last meaningful element (dir or file)
func getLastElementIndex(fileInfos []os.FileInfo, printFiles bool) int {
	if printFiles {
		return len(fileInfos) - 1
	}
	var lastElementIndex int
	for index := len(fileInfos) - 1; index >= 0; index-- {
		if fileInfos[index].IsDir() {
			lastElementIndex = index
			break
		}
	}
	return lastElementIndex
}

func recursiveTree(writer io.Writer, path string, printFiles bool, currectPrefix string) error {
	file, fileInfos, err := getFileInfos(path)
	defer file.Close()
	if err != nil {
		return err
	}

	lastElementIndex := getLastElementIndex(fileInfos, printFiles)

	// iterate over files and dirs
	for fileInfoIndex, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && !printFiles {
			continue
		}
		writer.Write(buildTreeLine(fileInfo, fileInfoIndex == lastElementIndex, currectPrefix))
		if fileInfo.IsDir() {
			if fileInfoIndex == lastElementIndex {
				recursiveTree(writer, filepath.Join(path, fileInfo.Name()), printFiles, currectPrefix+tab)
			} else {
				recursiveTree(writer, filepath.Join(path, fileInfo.Name()), printFiles, currectPrefix+verticalPrefix+tab)
			}
		}
	}
	return nil
}

func dirTree(writer io.Writer, path string, printFiles bool) error {
	return recursiveTree(writer, path, printFiles, "")
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

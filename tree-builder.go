package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Symlink struct {
	New string `json:"New"`
	Old string `json:"Old"`
}

type TreeNode struct {
	Directories map[string]*TreeNode `json:"Directories"`
	Files       []string             `json:"Files"`
	Symlinks    []Symlink            `json:"Symlinks"`
}

func parseTree(lines []string, root *TreeNode, level int) []string {
	for len(lines) > 0 {
		if len(lines) == 0 {
			return lines
		}

		strippedLine, hasLevel := checkLineLevel(lines[0], level)
		if !hasLevel {
			return lines
		}
		lines = lines[1:]

		isDirectory := false
		if len(lines) > 0 {
			_, nextLineHasLevel := checkLineLevel(lines[0], level+1)
			isDirectory = nextLineHasLevel
		}

		filename := ""
		if strings.HasPrefix(strippedLine, "├── ") {
			filename = strings.TrimPrefix(strippedLine, "├── ")
		} else if strings.HasPrefix(strippedLine, "└── ") {
			filename = strings.TrimPrefix(strippedLine, "└── ")
		} else {
			panic("Not an element: " + strippedLine)
		}

		if strings.Contains(filename, " -> ") {
			parts := strings.Split(filename, " -> ")
			if len(parts) == 2 {
				symlink := Symlink{
					New: parts[0],
					Old: parts[1],
				}
				root.Symlinks = append(root.Symlinks, symlink)
			} else {
				panic("Unexpected error parsing symlink: " + filename)
			}
		} else if isDirectory {
			branch := &TreeNode{
				Directories: make(map[string]*TreeNode),
				Files:       []string{},
			}
			lines = parseTree(lines, branch, level+1)
			root.Directories[filename] = branch
		} else {
			root.Files = append(root.Files, filename)
		}
	}
	return lines
}

func checkLineLevel(line string, level int) (string, bool) {
	prefixCount := 0
	for len(line) > 0 {
		if strings.HasPrefix(line, "│   ") {
			line = strings.TrimPrefix(line, "│   ")
			prefixCount++
		} else if strings.HasPrefix(line, "    ") {
			line = strings.TrimPrefix(line, "    ")
			prefixCount++
		} else {
			break
		}
	}
	return line, prefixCount == level
}

func isSummaryLine(line string) bool {
	return strings.Contains(line, "directories") && strings.Contains(line, "files")
}

// createStructure creates directories, files, and symlinks in the specified path
func createStructure(root *TreeNode, currentPath string) error {
	// Create directories
	for dirName, subDir := range root.Directories {
		dirPath := fmt.Sprintf("%s/%s", currentPath, dirName)
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory %s: %v", dirPath, err)
		}
		// Recurse into subdirectories
		err = createStructure(subDir, dirPath)
		if err != nil {
			return err
		}
	}

	// Create files
	for _, fileName := range root.Files {
		filePath := fmt.Sprintf("%s/%s", currentPath, fileName)
		_, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating file %s: %v", filePath, err)
		}
	}

	// Create symlinks
	for _, symlink := range root.Symlinks {
		newPath := fmt.Sprintf("%s/%s", currentPath, symlink.New)
		oldPath := fmt.Sprintf("%s/%s", currentPath, symlink.Old)

		// Create symlink
		err := os.Symlink(oldPath, newPath)
		if err != nil {
			return fmt.Errorf("error creating symlink from %s to %s: %v", newPath, oldPath, err)
		}
	}

	return nil
}

func main() {
	treeFile := "tree.txt"

	// Open the tree.txt file
	file, err := os.Open(treeFile)
	if err != nil {
		fmt.Printf("Error opening tree.txt: %v\n", err)
		return
	}
	defer file.Close()

	// Read all lines from the file
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip the root marker "." or summary lines
		if line == "." || isSummaryLine(line) || line == " " {
			continue
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading tree.txt: %v\n", err)
		return
	}

	// Parse the tree structure recursively
	root := &TreeNode{
		Directories: make(map[string]*TreeNode),
		Files:       []string{},
		Symlinks:    []Symlink{},
	}
	lines = parseTree(lines, root, 0)
	if len(lines) > 0 {
		fmt.Printf("%v\n", lines)
		panic("Lines left!")
	}

	// Create the output directory
	err = os.MkdirAll("output", 0755)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Create the structure in the output directory
	err = createStructure(root, "output")
	if err != nil {
		fmt.Printf("Error creating structure: %v\n", err)
		return
	}

	fmt.Println("Structure created successfully in 'output' directory")
}

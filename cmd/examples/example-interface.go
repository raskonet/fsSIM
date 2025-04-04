package main

import (
	"fmt"
	"io"
	"os"

	"github.com/raskonet/fsSIM/memfs"
)

func main() {
	// Create a new file system
	fs := memfs.NewFileSystem()

	// Create some directories
	err := fs.Mkdir("/usr/local/bin")
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Create a file
	file, err := fs.Create("/usr/local/bin/hello.txt")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}

	// Write to the file
	_, err = file.Write([]byte("Hello, World!"))
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}

	// Close the file
	err = file.Close()
	if err != nil {
		fmt.Printf("Error closing file: %v\n", err)
		os.Exit(1)
	}

	// Read the file
	data, err := fs.ReadFile("/usr/local/bin/hello.txt")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("File content: %s\n", string(data))

	// List directory contents
	nodes, err := fs.ReadDir("/usr/local/bin")
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Directory contents:")
	for _, node := range nodes {
		info, err := fs.Stat("/usr/local/bin/" + node.Name())
		if err != nil {
			fmt.Printf("Error getting file info: %v\n", err)
			continue
		}

		typeStr := "file"
		if info.IsDir {
			typeStr = "dir"
		}

		fmt.Printf("  %s (%s, %d bytes, created: %s)\n",
			info.Name, typeStr, info.Size, info.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Open and read a file using io.Reader
	file, err = fs.Open("/usr/local/bin/hello.txt")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}

	buf := make([]byte, 5)
	fmt.Println("Reading file in chunks:")
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Printf("Error reading file: %v\n", err)
			break
		}
		if n == 0 {
			break
		}
		fmt.Printf("  Read %d bytes: %s\n", n, string(buf[:n]))
	}

	// Seek back to beginning and read again
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		fmt.Printf("Error seeking: %v\n", err)
		os.Exit(1)
	}

	allData, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading all: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Read all: %s\n", string(allData))

	// Close the file
	err = file.Close()
	if err != nil {
		fmt.Printf("Error closing file: %v\n", err)
		os.Exit(1)
	}

	// Rename a file
	err = fs.Rename("/usr/local/bin/hello.txt", "/usr/local/bin/renamed.txt")
	if err != nil {
		fmt.Printf("Error renaming file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("File renamed successfully")

	// Remove a file
	err = fs.Remove("/usr/local/bin/renamed.txt")
	if err != nil {
		fmt.Printf("Error removing file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("File removed successfully")

	// Try to read a non-existent file
	_, err = fs.ReadFile("/usr/local/bin/hello.txt")
	if err != nil {
		fmt.Printf("Expected error reading non-existent file: %v\n", err)
		if memfs.IsNotExist(err) {
			fmt.Println("Verified error is a 'not exist' error")
		}
	}

	// Print number of open files
	fmt.Printf("Open files: %d\n", fs.GetOpenFileCount())
}

package utils

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

// TODO: add command for rename module name?

func UpdateModuleNameAndImports(projectPath, moduleName *string, warnings *[]*string) error {
	defaultModuleName := []byte("bykevin.work/refiber")
	newModuleName := []byte(*moduleName)

	var folderPathError error

	var once sync.Once
	handleErr := func(e error) {
		once.Do(func() {
			folderPathError = e
		})
	}

	targetFolders := []string{
		path.Join(*projectPath, "routes"),
		path.Join(*projectPath, "app"),
	}

	targetFilePathsCh := make(chan string)
	var finishedWalk atomic.Uint32
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for filePath := range targetFilePathsCh {
			mainGo, err := os.ReadFile(filePath)
			if err != nil {
				handleErr(err)
			}

			newMainGo := bytes.ReplaceAll(mainGo, defaultModuleName, newModuleName)
			if err := os.WriteFile(filePath, newMainGo, 0644); err != nil {
				handleErr(err)
			}
		}
	}()

	for _, folderPath := range targetFolders {
		wg.Add(1)
		go func(fp string) {
			defer wg.Done()

			err := filepath.Walk(
				fp,
				func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if info.IsDir() {
						return nil
					}

					if !strings.Contains(info.Name(), ".go") {
						return nil
					}

					targetFilePathsCh <- path

					return nil
				},
			)
			if err != nil {
				handleErr(err)
			}

			finishedWalk.Add(1)
			if finishedWalk.Load() == uint32(len(targetFolders)) {
				targetFilePathsCh <- filepath.Join(*projectPath, "go.mod")
				targetFilePathsCh <- filepath.Join(*projectPath, "main.go")

				close(targetFilePathsCh)
			}
		}(folderPath)
	}

	wg.Wait()

	if folderPathError != nil && warnings != nil {
		wr := folderPathError.Error()
		*warnings = append(*warnings, &wr)
	}

	return nil
}

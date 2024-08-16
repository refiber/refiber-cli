package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/refiber/refiber-cli/cmd/ui"
	"github.com/refiber/refiber-cli/cmd/ui/progress"
	"github.com/refiber/refiber-cli/cmd/ui/textInput"
	"github.com/refiber/refiber-cli/cmd/utils"
)

// TODO: add npm flag
// with npm flag we can directly install node modules and build it
// use this as reference https://github.com/charmbracelet/bubbletea/tree/master/examples/package-manager

// TODO: add module flag

var installerCmd = &cobra.Command{
	Use:   "new",
	Short: "Initiate a new Refiber Project",
	Long:  `Initiate a new Refiber Project`,
	Run:   installer,
}

var warnings []*string

func init() {
	rootCmd.AddCommand(installerCmd)
}

func installer(cmd *cobra.Command, args []string) {
	fmt.Println()

	var projectName string

	if len(args) < 1 {
		p := tea.NewProgram(textInput.InitialTextInputModel(&projectName, &textInput.Config{
			Header:      ui.TextTitle.Render("Please provide a project name"),
			Placeholder: "my-app",
			Validation: func(s string) error {
				matched, _ := regexp.Match("^[a-zA-Z0-9_-]+$", []byte(s))
				if !matched {
					return fmt.Errorf("Invalid project name")
				}
				return nil
			},
		}))
		if _, err := p.Run(); err != nil {
			cobra.CheckErr(ui.TextError.Render(err.Error()))
			return
		}
	} else {
		projectName = args[0]
	}

	if projectName == "" {
		fmt.Println(ui.TextWarning.Render("Project creation has been canceled"))
		fmt.Println()
		return
	}

	if projectName != "" && utils.DoesDirectoryExistAndIsNotEmpty(projectName) {
		cobra.CheckErr(ui.TextError.Render(fmt.Sprintf(`directory %s already exists and is not empty. Please choose a different name`, projectName)))
		return
	}
	projectName = strings.TrimSpace(projectName)

	var moduleName string
	p := tea.NewProgram(textInput.InitialTextInputModel(&moduleName, &textInput.Config{
		Header:      ui.TextTitle.Render("Please provide a module name"),
		Placeholder: "bykevin.work/refiber",
		Validation: func(s string) error {
			matched, _ := regexp.Match(`^([a-zA-Z0-9_\-\.]+\/)*[a-zA-Z0-9_\-\.]+$`, []byte(s))
			if !matched {
				return fmt.Errorf("Invalid module name")
			}
			return nil
		},
	}))
	if _, err := p.Run(); err != nil {
		cobra.CheckErr(ui.TextError.Render(err.Error()))
	}

	if moduleName == "" {
		fmt.Println(ui.TextWarning.Render("No module name provided. Using the default module name"))
		fmt.Println()
	}

	moduleName = strings.TrimSpace(moduleName)

	progressBar := tea.NewProgram(progress.InitialProgressModel("Preparing..."))
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if _, err := progressBar.Run(); err != nil {
			cobra.CheckErr(ui.TextError.Render(err.Error()))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := createNewProject(&projectName, &moduleName, progressBar); err != nil {
			cobra.CheckErr(ui.TextError.Render(err.Error()))
		}
	}()
	defer utils.DeferTeaPanicHandler(progressBar)

	wg.Wait()

	if releaseErr := progressBar.ReleaseTerminal(); releaseErr != nil {
		fmt.Printf("Problem releasing terminal: %v", releaseErr)
	}

	fmt.Println("  " + ui.TextGreen.Render("cd") + " " + ui.TextGray.Render(projectName))
	fmt.Println()
	fmt.Println("  " + ui.TextGreen.Render("npm") + " " + ui.TextGray.Render("i && ") + ui.TextGreen.Render("npm") + " " + ui.TextGray.Render("run build"))
	fmt.Println()
	fmt.Println("  " + ui.TextGreen.Render("air"))

	if len(warnings) > 0 {
		fmt.Println()
		for _, wr := range warnings {
			fmt.Println("  " + ui.TextWarning.Render(*wr))
		}
	}
}

func createNewProject(projectName, moduleName *string, progressBar *tea.Program) error {
	progressBar.Send(progress.ProgressMsg{Value: 0.0})

	currentWorkingDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// check and create project folder if not exist
	projectPath := filepath.Join(currentWorkingDir, *projectName)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		err := os.MkdirAll(projectPath, 0751)
		if err != nil {
			return err
		}
	}

	progressBar.Send(progress.ProgressMsg{Value: 0.25})

	/*
	 * I am not using the /latest endpoint because it will not work for pre-release labeled versions.
	 * For now, pre-release labeled versions will also be downloaded.
	 */
	refiberReleasesURL := "https://github.com/refiber/refiber/releases"

	resp, err := http.Get(refiberReleasesURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	progressBar.Send(progress.ProgressMsg{Value: 0.50})

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed when try to get releases data")
	}

	rawRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	respBody := string(rawRespBody)
	if respBody == "" {
		return fmt.Errorf("failed when try to get releases data")
	}

	// get latest release tag
	releases, err := utils.MatchAllStringByRegex(`<h2 class="sr-only" id.*v(.*)<\/h2>`, respBody)
	if err != nil || len(releases) == 0 {
		if err != nil {
			return err
		}

		return fmt.Errorf("could not find the latest release tag version")
	}
	latestRelease := releases[0]
	refiberDownloadArchiveURL := fmt.Sprintf(`https://github.com/refiber/refiber/archive/refs/tags/v%s.tar.gz`, *latestRelease)

	progressBar.Send(progress.ProgressMsg{Value: 0.60})

	// download and extract release tag to project folder
	tempFileName := fmt.Sprintf("%v_v%s.tar.gz", time.Now().Unix(), *latestRelease)
	tempFilePath := filepath.Join(currentWorkingDir, tempFileName)
	if err = utils.DownloadFile(refiberDownloadArchiveURL, tempFilePath); err != nil {
		return err
	}
	progressBar.Send(progress.ProgressMsg{Value: 0.80})
	if err = extractTarGz(tempFilePath, projectPath, "refiber-"+*latestRelease); err != nil {
		return err
	}
	if err = os.Remove(tempFilePath); err != nil {
		return err
	}

	progressBar.Send(progress.ProgressMsg{Value: 0.90})
	if *moduleName != "" {
		if err = utils.UpdateModuleNameAndImports(&projectPath, moduleName, &warnings); err != nil {
			return err
		}
	}

	// copy .env.example to .env
	utils.CopyFile(filepath.Join(projectPath, ".env.example"), filepath.Join(projectPath, ".env"))

	progressBar.Send(progress.ProgressMsg{Value: 1.0})

	return nil
}

func extractTarGz(filePath, destPath, folderName string) error {
	// Open the tar.gz file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Check if the current file is within the desired folder
		if strings.HasPrefix(header.Name, folderName+"/") {
			// Determine the relative path within the folder
			relPath := strings.TrimPrefix(header.Name, folderName+"/")

			// Determine the path for the current file
			target := filepath.Join(destPath, relPath)

			// Ensure that the target path is within the destination directory
			rel, err := filepath.Rel(destPath, target)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
			if strings.HasPrefix(rel, "..") {
				return fmt.Errorf("illegal file path: %s", target)
			}

			// Create directories as needed
			if header.FileInfo().IsDir() {
				if err := os.MkdirAll(target, os.ModePerm); err != nil {
					return fmt.Errorf("failed to create directory: %w", err)
				}
				continue
			}

			// Create the file
			file, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer file.Close()

			// Write the file contents
			if _, err := io.Copy(file, tarReader); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			// Set permissions on the file
			if err := file.Chmod(header.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to set file permissions: %w", err)
			}
		}
	}

	return nil
}

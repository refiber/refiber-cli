package utils

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

func MatchAllStringByRegex(regex, str string) ([]*string, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	result := r.FindAllSubmatch([]byte(str), -1)

	if len(result) < 1 {
		return nil, nil
	}

	var data []*string

	for _, v := range result {
		d := string(v[1])
		data = append(data, &d)
	}

	return data, nil
}

func DownloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func DoesDirectoryExistAndIsNotEmpty(name string) bool {
	if _, err := os.Stat(name); err == nil {
		dirEntries, err := os.ReadDir(name)
		if err != nil {
			return false
		}
		if len(dirEntries) > 0 {
			return true
		}
	}
	return false
}

func CopyFile(sourcePath, destinationPath string) error {
	// Open the source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	// Optionally, you can also flush the destination file to ensure all data is written
	err = destinationFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

// ExecuteCmd provides a shorthand way to run a shell command
func ExecuteCmd(name string, args []string, dir string) error {
	command := exec.Command(name, args...)
	command.Dir = dir
	var out bytes.Buffer
	command.Stdout = &out
	if err := command.Run(); err != nil {
		return err
	}
	return nil
}

func DeferTeaPanicHandler(p *tea.Program) {
	if r := recover(); r != nil {
		fmt.Println("The program encountered an unexpected issue and had to exit. The error was:", r)
		fmt.Println("If you continue to experience this issue, please post a message on our GitHub page.")
		fmt.Println()
		if releaseErr := p.ReleaseTerminal(); releaseErr != nil {
			fmt.Printf("Problem releasing terminal: %v", releaseErr)
		}
	}
}

func DoesDirectoryOrFileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func GetRefiberTemplateDirPath(currentWorkingDir *string) (*string, error) {
	// check is current dir is a refiber project by checking go.mod file
	goModFilePath := filepath.Join(*currentWorkingDir, "go.mod")
	goModFileContent, err := os.ReadFile(goModFilePath)
	if err != nil {
		return nil, fmt.Errorf("the current folder path is not inside the Refiber project")
	}

	var refiberVersion string

	pattern := `github\.com/refiber/framework\s(v.+)`
	match := regexp.MustCompile(pattern).FindSubmatch(goModFileContent)
	if len(match) > 1 {
		refiberVersion = string(match[1])
	} else {
		return nil, fmt.Errorf("the current folder path is not inside the Refiber project")
	}

	templatePathInVendor := filepath.Join(*currentWorkingDir, "vendor", "github.com", "refiber", "framework", "templates")
	if DoesDirectoryOrFileExist(templatePathInVendor) {
		return &templatePathInVendor, nil
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	templatePathInGoPkg := filepath.Join(gopath, "pkg", "mod", "github.com", "refiber", fmt.Sprintf("framework@%s", refiberVersion), "templates")
	if DoesDirectoryOrFileExist(templatePathInGoPkg) {
		return &templatePathInGoPkg, nil
	}

	return nil, fmt.Errorf("template folder not found. Make sure you are using the newest version of Refiber")
}

func ExecuteTemplate(t *template.Template, data interface{}) ([]byte, error) {
	var contentBuf bytes.Buffer

	if err := t.Execute(&contentBuf, data); err != nil {
		return nil, err
	}

	return contentBuf.Bytes(), nil
}

func WriteFile(filename, path string, content []byte) error {
	file, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func ListFolders(dirPath string) (*[]*string, error) {
	if dirPath == "" {
		return nil, fmt.Errorf("unable to list folder. the provided path is invalid")
	}

	var folders []*string

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			f := file.Name()
			folders = append(folders, &f)
		}
	}

	return &folders, nil
}

func GetLowercaseFirstChar(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func GetLastPathName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

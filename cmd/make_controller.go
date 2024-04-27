package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/refiber/refiber-cli/cmd/ui"
	"github.com/refiber/refiber-cli/cmd/ui/selectInput"
	"github.com/refiber/refiber-cli/cmd/ui/textInput"
	"github.com/refiber/refiber-cli/cmd/utils"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var makeControllerCmd = &cobra.Command{
	Use:   "make:controller",
	Short: "Generate a Controller file",
	Run:   generateController,
}

func init() {
	rootCmd.AddCommand(makeControllerCmd)
	makeControllerCmd.Flags().BoolP("crud", "c", false, "Create CRUD controller")
}

func generateController(cmd *cobra.Command, args []string) {
	fmt.Println()

	var input string
	if len(args) < 1 {
		p := tea.NewProgram(textInput.InitialTextInputModel(&input, "Please provide a controller name"))
		if _, err := p.Run(); err != nil {
			cobra.CheckErr(ui.TextError.Render(err.Error()))
		}
	} else {
		input = args[0]
	}

	currentWorkingDir, err := os.Getwd()
	if err != nil {
		cobra.CheckErr(ui.TextError.Render(err.Error()))
	}

	templateDirPath, err := utils.GetRefiberTemplateDirPath(&currentWorkingDir)
	if err != nil {
		cobra.CheckErr(ui.TextError.Render(err.Error()))
	}

	// check target folder
	cName, cDirPath, err := getControllerNameAndPath(input, &currentWorkingDir)
	if err != nil {
		cobra.CheckErr(ui.TextError.Render(err.Error()))
	}

	var packageName string
	if cDirPath == nil {
		// check folders in controllers, if there are more then one user should choose one
		availableControllerPathFolders, err := utils.ListFolders(filepath.Join(currentWorkingDir, "app", "controllers"))
		if err != nil {
			cobra.CheckErr(ui.TextError.Render(err.Error()))
		}

		aCPFCount := len(*availableControllerPathFolders)

		if aCPFCount < 1 {
			cobra.CheckErr(ui.TextError.Render("no controller folder found in your app/controller"))
		} else if aCPFCount > 1 {
			p := tea.NewProgram(selectInput.InitialSelectInputModel(&packageName, "Select the folder where you will save the controller", *availableControllerPathFolders))
			_, err := p.Run()
			if err != nil {
				cobra.CheckErr(ui.TextError.Render(err.Error()))
			}
		} else {
			n := *availableControllerPathFolders
			packageName = *n[0]
		}
	} else {
		// use last folder path name as packageName
		packageName = utils.GetLastPathName(*cDirPath)
	}

	if cDirPath == nil && packageName != "" {
		cdp := filepath.Join(*getControllersDirPath(&currentWorkingDir), packageName)
		cDirPath = &cdp
	}

	// verify if the controller already exists
	if utils.DoesDirectoryOrFileExist(filepath.Join(*cDirPath, *cName+".go")) {
		cobra.CheckErr(ui.TextError.Render("the " + *cName + ".go" + " already exist"))
	}

	templateFileName := "controller.go.tmpl"

	useCrud, _ := cmd.Flags().GetBool("crud")
	if useCrud {
		templateFileName = "controller_crud.go.tmpl"
	}

	// get template file and content
	cTemplatePath := filepath.Join(*templateDirPath, "controller", templateFileName)
	cTemplateContent, err := os.ReadFile(cTemplatePath)
	if err != nil {
		cobra.CheckErr(ui.TextError.Render(err.Error()))
	}

	tmpl, err := template.New(*cName).Parse(string(cTemplateContent))
	if err != nil {
		cobra.CheckErr(ui.TextError.Render(err.Error()))
	}

	type ControllerData struct {
		PackageName    string // web
		MethodName     string // Product
		ControllerName string // ProductController
		ModelName      string // productController
		ReciverName    string // c
	}

	modelName := utils.GetLowercaseFirstChar(*cName)

	data := &ControllerData{
		PackageName:    packageName,
		MethodName:     strings.ReplaceAll(*cName, "Controller", ""),
		ControllerName: *cName,
		ModelName:      modelName,
		ReciverName:    "c",
	}

	// inject data to the template
	buf, err := utils.ExecuteTemplate(tmpl, data)
	if err != nil {
		cobra.CheckErr(ui.TextError.Render(err.Error()))
	}

	// write template file
	if err = utils.WriteFile(data.ControllerName+".go", *cDirPath, buf); err != nil {
		cobra.CheckErr(ui.TextError.Render(err.Error()))
	}

	fmt.Println(ui.TextGreen.Render(*cName + ".go successfully created!"))
}

func getControllersDirPath(currentWorkingDir *string) *string {
	if currentWorkingDir == nil {
		return nil
	}

	p := filepath.Join(*currentWorkingDir, "app", "controllers")
	return &p
}

func getControllerNameAndPath(input string, currentWorkingDir *string) (name, path *string, err error) {
	// Split the input string by "/"
	parts := strings.Split(input, "/")

	// Extract the name from the last part
	n := cases.Title(language.English).String(parts[len(parts)-1])
	n = strings.ReplaceAll(n, "Controller", "")
	n = strings.ReplaceAll(n, "controller", "")
	n = strings.ReplaceAll(n, ".go", "")
	n = strings.ReplaceAll(n, ".", "")
	n = n + "Controller"
	name = &n

	if len(parts) > 1 {
		// Extract the path from the remaining parts
		p := filepath.Join(*getControllersDirPath(currentWorkingDir), filepath.Join(parts[:len(parts)-1]...))
		path = &p

		// Check if the path exists, if not create it
		if _, err := os.Stat(*path); os.IsNotExist(err) {
			if err := os.MkdirAll(*path, 0755); err != nil {
				return nil, nil, err
			}
		}
	}

	return name, path, err
}

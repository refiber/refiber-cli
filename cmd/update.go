package cmd

import (
	"fmt"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/refiber/refiber-cli/cmd/ui"
	"github.com/refiber/refiber-cli/cmd/ui/spinner"
	"github.com/refiber/refiber-cli/cmd/utils"
	"github.com/spf13/cobra"
)

/**
 * TODO: check latest version and print it, but if possible
 * if current version is the same with the latest then print it
 */

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update refiber cli",
	Run:   update,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func update(cmd *cobra.Command, args []string) {
	fmt.Println()

	spinner := tea.NewProgram(spinner.InitialSpinnerModel("updating..."))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if _, err := spinner.Run(); err != nil {
			cobra.CheckErr(ui.TextError.Render(err.Error()))
		}
	}()
	defer utils.DeferTeaPanicHandler(spinner)

	if err := installLatestCLI(); err != nil {
		cobra.CheckErr(ui.TextError.Render(err.Error()))
	}

	if releaseErr := spinner.ReleaseTerminal(); releaseErr != nil {
		fmt.Printf("Problem releasing terminal: %v", releaseErr)
	}

	fmt.Println(ui.TextGreen.PaddingLeft(1).Render("update complete!"))
}

func installLatestCLI() error {
	currentWorkingDir, err := os.Getwd()

	if err != nil {
		return err
	}

	sourceURL := "github.com/refiber/refiber-cli@latest"
	if err = utils.ExecuteCmd("go", []string{"install", sourceURL}, currentWorkingDir); err != nil {
		return err
	}

	return nil
}

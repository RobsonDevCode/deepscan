package cmd

import (
	"fmt"

	"github.com/fatih/color"

	setupservice "github.com/RobsonDevCode/deepscan/internal/services/setupService"
	"github.com/spf13/cobra"
)

var setUpCmd = &cobra.Command{
	Use:   "setup [repository portal url]",
	Short: "setup repository so we can begin scanning projects",
	Long: `setup creates a connection to the given repository portal
		   This allows the 'scan' command to scan specific or all projects within for vulnerabilities`,
	Args: cobra.MaximumNArgs(2),
	RunE: runSetUp,
}

const (
	UrlFlag     = "url"
	ProfileFlag = "profile"
)

func runSetUp(cmd *cobra.Command, urls []string) error {
	url, _ := cmd.Flags().GetString(UrlFlag)
	profile, _ := cmd.Flags().GetString(ProfileFlag)

	fmt.Print("\n Setting up scanner...")

	if err := setupservice.CreateSetupFile(url, profile); err != nil {
		return err
	}

	fmt.Print(color.GreenString("\n Scanner set up, please run scan command to scan projects!"))
	return nil
}

func init() {
	setUpCmd.Flags().StringP("url", "u", "", "Url used to connect to your org's repository.")
	setUpCmd.Flags().StringP("profile", "p", "", "Profile or Project(on azure) that your project repositories are listed under.")

	rootCmd.AddCommand(setUpCmd)
}

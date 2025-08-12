package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fatih/color"

	supportedproviders "github.com/RobsonDevCode/deepscan/internal/constants/supportedProviders"
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
	UrlFlag      = "url"
	AccountFlag  = "account"
	ProviderFlag = "provider"
)

func runSetUp(cmd *cobra.Command, urls []string) error {
	fmt.Print("\n Setting up scanner...")

	provider, _ := cmd.Flags().GetString(ProviderFlag)
	orgUrl, _ := cmd.Flags().GetString(UrlFlag)
	profile, _ := cmd.Flags().GetString(AccountFlag)

	if provider == supportedproviders.Azure && orgUrl == "" {
		scanner := bufio.NewScanner(os.Stdin)

		fmt.Printf("\n %s", color.HiMagentaString("\nPlease Enter Azure Org Url e.g. https://dev.azure.com/your_org/: "))
		if scanner.Scan() {
			orgUrl = scanner.Text()
		}
	}

	if err := setupservice.CreateSetupFile(orgUrl, provider, profile); err != nil {
		return err
	}

	fmt.Print(color.GreenString("\n Scanner set up, please run scan command to scan projects!"))
	return nil
}

func init() {
	setUpCmd.Flags().StringP("provider", "p", "", "Provider that the repository is saved on e.g. Github Or Azure Devops")
	setUpCmd.MarkFlagRequired("provider")

	setUpCmd.Flags().StringP("url", "u", "", "Url used to connect to your org's repository.")

	setUpCmd.Flags().StringP("account", "a", "", "Profile or Project(on azure) that your project repositories are listed under.")
	setUpCmd.MarkFlagRequired("account")

	rootCmd.AddCommand(setUpCmd)
}

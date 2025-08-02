package cmd

import (
	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	"github.com/RobsonDevCode/deepscan/internal/constants/exportExcelOptions"
	excelexportservice "github.com/RobsonDevCode/deepscan/internal/services/excelExportService"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [project-name-or-path]",
	Short: "scan project for dependency vulnerabilities",
	Long: `scan project for dependency vulnerabilities.

		   If no argument is provided, lists all available projects to scan.
           If a project name or path is provided, scans that specific project.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

var allFlag bool

func runScan(cmd *cobra.Command, projects []string) error {
	ctx := cmd.Context()

	var scannedPackages []models.ScannedPackage
	if allFlag {
		scannerResponse, err := scannerSelectionService.ScanAll(ctx)
		if err != nil {
			return err
		}
		scannedPackages = scannerResponse
	} else {
		scannerResponse, err := scannerSelectionService.Scan(cmd, ctx)
		if err != nil {
			return err
		}
		scannedPackages = scannerResponse
	}

	if len(scannedPackages) == 0 {
		return nil
	}

	choice, err := excelexportservice.SelectExportPackagesToExcel()
	if err != nil {
		return err
	}

	if choice == exportExcelOptions.Yes {
		if err := excelexportservice.ExportPackageTable(scannedPackages, allFlag); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	scanCmd.Flags().StringP("dir", "d", "", "Process from local directory(most effiecent)")
	scanCmd.Flags().StringP("ssh", "s", "", "Processes using the ssh url for the project repository")
	scanCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Scans all projects for package vulnerabilities")

	rootCmd.AddCommand(scanCmd)
}

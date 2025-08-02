package tablewriterservice

import (
	"fmt"
	"os"
	"slices"

	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	"github.com/RobsonDevCode/deepscan/internal/constants/tableHeaders"
	"github.com/RobsonDevCode/deepscan/internal/extensions"
	frameworkconstants "github.com/RobsonDevCode/deepscan/internal/scanner/constants/framework"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func DisplayPackagesTable(packages []models.ScannedPackage) {
	if len(packages) == 0 {
		fmt.Print(color.GreenString("\n No Package Vulnerabilities!\n"))
		return
	}

	fmt.Printf("\nHave %d at Display", len(packages))
	slices.SortFunc(packages, func(a, b models.ScannedPackage) int {
		return b.RiskScore - a.RiskScore
	})

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}},
		})),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoWrap:  tw.WrapNormal,
					MergeMode: tw.MergeHierarchical}, //wrap long content like summary and discription
				Alignment:    tw.CellAlignment{Global: tw.AlignCenter},
				ColMaxWidths: tw.CellWidth{Global: 10},
			},
		}),
	)

	table.Header(tableHeaders.ExcelPackageTableHeaders)

	vulnerablityCount := 0
	for _, pkg := range packages {
		for i := range pkg.Vulnerabilities {
			vulnerablityCount++
			table.Append([]string{
				pkg.ServiceName,
				"",
				pkg.Vulnerabilities[i].Package.Name,
				pkg.Vulnerabilities[i].CurrentVersion,
				extensions.TruncateString(pkg.Summary, 50),
				extensions.TruncateString(pkg.Description, 50),
				pkg.Severity,
				pkg.Vulnerabilities[i].FirstPatchedVersion,
				pkg.GithubReviewedAt.Format("2006-01-02"),
			})
		}

	}

	fmt.Printf("\n Found %d Package Vulnerabilities: \n", vulnerablityCount)
	fmt.Printf("\n%s\n", color.HiMagentaString("Download table to see full results"))

	table.Render()
}

func DisplayFailedScanTable(failedScans []models.FailedProjectScan) {
	if len(failedScans) == 0 {
		return
	}

	fmt.Printf("%s", color.RedString("\nFailed To Process Projects: \n"))
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}},
		})),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoWrap:  tw.WrapNormal,
					MergeMode: tw.MergeHierarchical}, //wrap long content like summary and discription
				Alignment:    tw.CellAlignment{Global: tw.AlignCenter},
				ColMaxWidths: tw.CellWidth{Global: 10},
			},
		}),
	)

	table.Header([]string{"Service Name", "Project Name", "Error"})

	for _, failedScan := range failedScans {

		table.Append([]string{
			failedScan.ServiceName,
			failedScan.ProjectName,
			extensions.TruncateString(failedScan.Error.Error(), 500),
		})
	}

	table.Render()
}

func DisplayInfomationTable(scannedProjects []models.ScannerResponse) {
	fmt.Print("\n Project Infomation: \n")

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}},
		})),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoWrap:  tw.WrapNormal,
					MergeMode: tw.MergeHierarchical}, //wrap long content like summary and discription
				Alignment:    tw.CellAlignment{Global: tw.AlignCenter},
				ColMaxWidths: tw.CellWidth{Global: 10},
			},
		}),
	)

	for _, project := range scannedProjects {
		needsUpgrade := "False"
		if project.Framework != frameworkconstants.LatestMaintainableFramework {
			needsUpgrade = "True" // has to be string so we can represent it the table
		}

		table.Header([]string{"Project", "Framework", "NeedsUpdating"})
		table.Append([]string{
			project.Name,
			project.Framework,
			needsUpgrade})
	}

	table.Render()
}

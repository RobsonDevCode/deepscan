package excelexportservice

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	"github.com/RobsonDevCode/deepscan/internal/constants/exportExcelOptions"
	"github.com/RobsonDevCode/deepscan/internal/constants/tableHeaders"
	"github.com/xuri/excelize/v2"
)

const saveFileTo = "./export"
const packageSheetName = "Package Vulnerabilities"

func ExportPackageTable(packages []models.ScannedPackage, isFullScan bool) error {
	if err := os.MkdirAll(saveFileTo, 0755); err != nil {
		return fmt.Errorf("error creating directory %s, %w", saveFileTo, err)
	}

	file := excelize.NewFile()
	defer file.Close()

	file.SetSheetName("Sheet1", packageSheetName)
	for i, header := range tableHeaders.ExcelPackageTableHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		file.SetCellValue(packageSheetName, cell, header)
	}

	//excelName identifier
	var name string
	for i, pkg := range packages {
		row := i + 2 // excel is 1 index and skip headers

		if name == "" && !isFullScan {
			name = pkg.ServiceName
		}

		for _, vuln := range pkg.Vulnerabilities {
			rowData := []interface{}{
				pkg.ServiceName,
				pkg.ProjectName,
				vuln.Package.Name,
				vuln.CurrentVersion,
				pkg.Summary,
				pkg.Description,
				vuln.FirstPatchedVersion,
				pkg.Severity,
				pkg.GithubReviewedAt.Format("2006-01-02"),
			}

			file.SetSheetRow(packageSheetName, fmt.Sprintf("A%d", row), &rowData)
			row++
		}
	}

	if isFullScan {
		name = "full_scan"
	}

	fileName := fmt.Sprintf("package_%s_vun_%s.xlsx", name, time.Now().Format("2006-01-02T15-04-05"))
	fullPath := filepath.Join(saveFileTo, fileName)

	if err := file.SaveAs(fullPath); err != nil {
		return fmt.Errorf("failed to save excel to %s, %w", fullPath, err)
	}

	fmt.Printf("Your file has been saved to: %s", fullPath)

	return nil
}

func SelectExportPackagesToExcel() (string, error) {

	prompt := &survey.Select{
		Message: "Export Package Table",
		Options: exportExcelOptions.ExcelOptions,
	}

	var selectedIndex int
	err := survey.AskOne(prompt, &selectedIndex)
	if err != nil {
		fmt.Print("selection cancelled")
		return "", fmt.Errorf("selection error: %w", err)
	}

	return exportExcelOptions.ExcelOptions[selectedIndex], nil
}

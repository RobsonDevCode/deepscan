package main

import (
	"fmt"

	"github.com/RobsonDevCode/deepscan/cmd"
	cache "github.com/RobsonDevCode/deepscan/internal/caching"
	client "github.com/RobsonDevCode/deepscan/internal/clients"
	"github.com/RobsonDevCode/deepscan/internal/configuration"
	scanner "github.com/RobsonDevCode/deepscan/internal/scanner"
	packagereaderservice "github.com/RobsonDevCode/deepscan/internal/services/packageReaderService"
	repositoryreaderservice "github.com/RobsonDevCode/deepscan/internal/services/repositoryReaderService"
	scanfileservice "github.com/RobsonDevCode/deepscan/internal/services/scanFileService"
	scansshservice "github.com/RobsonDevCode/deepscan/internal/services/scanShhService"
	scannerselectionservice "github.com/RobsonDevCode/deepscan/internal/services/scannerSelectionService"
	azurecommandExcecutor "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/azureCommands"
)

func main() {

	cacheIntance := cache.Cache{}
	azureCommandExcecutor := azurecommandExcecutor.NewAzureCommandExcecutor(&cacheIntance)
	config, err := configuration.Load()
	if err != nil {
		fmt.Printf("error staring command line: %s", err.Error())
		return
	}

	githubClient, err := client.NewGithubClient(config, &cacheIntance)
	if err != nil {
		fmt.Printf("error staring command line: %s", err.Error())
		return
	}

	packageReader := packagereaderservice.NewPackageReader()
	scanner := scanner.NewScanner(githubClient, packageReader)
	repositoryReader := repositoryreaderservice.NewRepositoryReaderService(azureCommandExcecutor)
	sshService := scansshservice.NewSshProcessor(scanner, &repositoryReader)
	fileService := scanfileservice.NewFileScannerService(scanner, packageReader)
	scanSelection := scannerselectionservice.NewScanSelection(sshService, fileService, &repositoryReader)

	// cant DI directly into the command so we use a setter
	cmd.SetScanSelection(scanSelection)
	cmd.Execute()
}

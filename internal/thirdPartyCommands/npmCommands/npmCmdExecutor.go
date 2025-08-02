package npmcommands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	npmmodels "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/models/npm"
)

func GetPackages(path string, ctx context.Context) (npmmodels.NpmPackageResponse, error) {
	cmd := exec.CommandContext(ctx, "npm", "ls", "--all", "--package-lock-only", "--json")

	cmd.Dir = path
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if len(output) == 0 {
				return npmmodels.NpmPackageResponse{}, fmt.Errorf("npm list failed in %s: %v", path, exitError)
			}
		} else {
			return npmmodels.NpmPackageResponse{}, fmt.Errorf("failed to run npm list in %s: %v", path, err)
		}
	}

	var result npmmodels.NpmPackageResponse
	if err := json.Unmarshal(output, &result); err != nil {
		return npmmodels.NpmPackageResponse{}, fmt.Errorf("failed to parse data from npm list output in %s: %w", path, err)
	}

	return result, nil
}

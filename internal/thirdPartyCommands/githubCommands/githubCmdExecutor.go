package githubcommands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	scannerconstants "github.com/RobsonDevCode/deepscan/internal/scanner/constants"
	"golang.org/x/sync/errgroup"
)

func CloneRepository(sshUrl string, ctx context.Context) error {
	if err := createTempDir(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "git", "clone", sshUrl)
	cmd.Dir = scannerconstants.TempDirctory
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout attempting to clone: %s", sshUrl)
		}

		return err
	}

	return nil
}

func CloneAll(urls []string, ctx context.Context) error {
	if err := createTempDir(); err != nil {
		return fmt.Errorf("\n error creating temp file: %w", err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	for _, url := range urls {
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return fmt.Errorf("context cancelled before starting clone of %s: %w", url, gCtx.Err())

			default:
				cmd := exec.CommandContext(gCtx, "git", "clone", url)
				cmd.Dir = scannerconstants.TempDirctory

				//this is needed as this will deadlock on git's hang time if not
				cmd.Env = append(os.Environ(),
					"GIT_TERMINAL_PROMPT=0", // Disable terminal prompts
					"GIT_ASKPASS=echo",      // Provide dummy askpass
					"SSH_ASKPASS=echo",      // Disable SSH prompts
				)

				if err := cmd.Run(); err != nil {
					if ctx.Err() == context.DeadlineExceeded {
						return fmt.Errorf("timeout attempting to clone: %s", urls)
					}

					//due to how the repo is set up, it throws an auth error for concurrent access
					//it still clones but we want to ignore auth erros for this repo
					if strings.Contains(err.Error(), "status 128") {
						//do nothing
					} else {
						return fmt.Errorf("failed to clone %s:  %w", url, err)
					}

				}

				fmt.Printf("\nclone completed for project: %s", url)
				return nil
			}
		})

	}

	if err := g.Wait(); err != nil {
		os.RemoveAll(scannerconstants.TempDirctory)
		return err
	}

	fmt.Printf("\n Successfully cloned all repos ")
	return nil
}

func createTempDir() error {
	if _, err := os.Stat(scannerconstants.TempDirctory); err == nil {
		os.RemoveAll(scannerconstants.TempDirctory)
	}

	if err := os.Mkdir(scannerconstants.TempDirctory, 0755); err != nil {
		return fmt.Errorf("error making temp directory: %w", err)
	}

	return nil
}

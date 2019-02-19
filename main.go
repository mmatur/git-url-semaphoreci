package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/containous/flaeg"
	"github.com/google/go-github/github"
	"github.com/ogier/pflag"
	"golang.org/x/oauth2"
)

// Config holds configuration.
type Config struct {
	Owner       string `description:"Repository owner"`
	Repo        string `description:"Repository name"`
	GithubToken string `description:"Github Token"`
	BranchName  string `description:"Branch name to check Token"`
	Label       string `description:"Label required to get correct ssh-url"`
}

// NoOption empty struct.
type NoOption struct{}

func main() {
	defaultCfg := &Config{}
	defaultPointerCfg := &Config{}

	rootCmd := &flaeg.Command{
		Name:                  "git-url-semaphoreci",
		Description:           "Git URL SemaphoreCI",
		Config:                defaultCfg,
		DefaultPointersConfig: defaultPointerCfg,
		Run: func() error {
			output, err := rootRun(defaultCfg)
			if err != nil {
				// if there is an error the output need to be empty
				output = ""
			}

			fmt.Print(output)
			return nil
		},
	}

	flag := flaeg.New(rootCmd, os.Args[1:])

	// version
	versionCmd := &flaeg.Command{
		Name:                  "version",
		Description:           "Display the version.",
		Config:                &NoOption{},
		DefaultPointersConfig: &NoOption{},
		Run: func() error {
			DisplayVersion()
			return nil
		},
	}

	flag.AddCommand(versionCmd)

	// Run command
	if err := flag.Run(); err != nil && err != pflag.ErrHelp {
		log.Printf("Error: %v\n", err)
	}
}

func rootRun(config *Config) (string, error) {
	if config.GithubToken == "" {
		config.GithubToken = os.Getenv("GITHUB_TOKEN")
	}

	if config.BranchName == "" {
		config.BranchName = os.Getenv("SEMAPHORE_GIT_BRANCH")
	}

	if err := validate(config); err != nil {
		return "", err
	}

	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GithubToken},
	))
	client := github.NewClient(tc)

	if strings.Contains(config.BranchName, "pull-request-") {
		s := strings.Split(config.BranchName, "pull-request-")
		if len(s) == 2 {
			ID, err := strconv.Atoi(s[1])
			if err != nil {
				return "", err
			}
			pr, _, err := client.PullRequests.Get(ctx, config.Owner, config.Repo, ID)
			if err != nil {
				return "", err
			}

			if config.Label != "" && !labelsContains(config.Label, pr.Labels) {
				return "", fmt.Errorf("PR labels does not contain %s", config.Label)
			}

			if pr.GetHead() != nil && pr.GetHead().GetRepo() != nil {
				return pr.GetHead().GetRepo().GetSSHURL(), nil
			}

			return "", fmt.Errorf("unable to get head of PR %d", ID)
		}
	}

	return os.Getenv("SEMAPHORE_GIT_URL"), nil
}

func labelsContains(label string, labels []*github.Label) bool {
	for _, value := range labels {
		if value.GetName() == label {
			return true
		}
	}

	return false
}

func validate(config *Config) error {
	if err := required(config.Owner, "owner"); err != nil {
		return err
	}

	if err := required(config.Repo, "repo"); err != nil {
		return err
	}

	if err := required(config.BranchName, "branchname"); err != nil {
		return err
	}

	return required(config.GithubToken, "githubtoken")
}

func required(field string, fieldName string) error {
	if len(field) == 0 {
		return fmt.Errorf("option %s is mandatory", fieldName)
	}
	return nil
}

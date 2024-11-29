// Copyright 2013-2014 Canonical Ltd.

// godeb dynamically translates stock upstream Go tarballs to deb packages.
//
// For details of how this tool works and context for why it was built,
// please refer to the following blog post:
//
//	http://blog.labix.org/2013/06/15/in-flight-deb-packages-of-go
package main

import (
	"encoding/json"
	"fmt"
	"go/build"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	GOARCH = build.Default.GOARCH
	GOOS   = build.Default.GOOS
)

func main() {
	if GOARCH == "arm" {
		GOARCH = "armv6l"
	}

	listCmd.Flags().BoolVarP(&includeAll, "all", "a", false, "Include all versions")

	rootCmd.AddCommand(listCmd, downloadCmd, installCmd, removeCmd)
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "godeb",
	Short: "godeb dynamically translates stock upstream Go tarballs to deb packages",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},

	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
}

var includeAll bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available Go versions",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		tbs, err := tarballs(includeAll)
		if err != nil {
			return err
		}
		for _, tb := range tbs {
			fmt.Println(tb.Version)
		}
		return nil
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove the installed Go package",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		args := []string{"dpkg", "--purge", "go"}
		if os.Getuid() != 0 {
			args = append([]string{"sudo"}, args...)
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("while removing go package: %w", err)
		}
		return nil
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download [version]",
	Short: "Download the Go package and transform it into a deb package",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		version := ""
		if len(args) == 1 {
			version = args[0]
		}
		return actionCommand(version, false)
	},
}

var installCmd = &cobra.Command{
	Use:   "install [version]",
	Short: "Download the Go package, transform it into a deb package, and install it",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		version := ""
		if len(args) == 1 {
			version = args[0]
		}
		return actionCommand(version, true)
	},
}

func actionCommand(version string, install bool) error {
	tbs, err := tarballs(true)
	if err != nil {
		return err
	}
	var url string
	if version == "" {
		version = tbs[0].Version
		url = tbs[0].URL
	} else {
		for _, tb := range tbs {
			if version == tb.Version {
				url = tb.URL
				break
			}
		}
	}

	installed, err := installedDebVersion()
	if err == errNotInstalled {
		// that's okay
	} else if err != nil {
		return err
	} else if install && debVersion(version) == installed {
		return fmt.Errorf("go version %s is already installed", version)
	}

	fmt.Println("processing", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", url, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("got status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	debName := fmt.Sprintf("go_%s_%s.deb", debVersion(version), debArch())
	deb, err := os.Create(debName + ".inprogress")
	if err != nil {
		return fmt.Errorf("cannot create deb: %v", err)
	}
	defer deb.Close()

	if err := createDeb(version, resp.Body, deb); err != nil {
		return err
	}
	if err := os.Rename(debName+".inprogress", debName); err != nil {
		return err
	}
	fmt.Println("package", debName, "ready")

	if install {
		args := []string{"dpkg", "-i", debName}
		if os.Getuid() != 0 {
			args = append([]string{"sudo"}, args...)
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("while installing go package: %v", err)
		}
	}
	return nil
}

type Tarball struct {
	URL     string
	Version string
}

type GolangDlFile struct {
	Arch     string `json:"arch"`
	Filename string `json:"filename"`
	Os       string `json:"os"`
	Version  string `json:"version"`
}

type GolangDlVersion struct {
	Version string         `json:"version"`
	Files   []GolangDlFile `json:"files"`
}

// REST API described in https://github.com/golang/website/blob/master/internal/dl/dl.go
func tarballs(includeAll bool) ([]*Tarball, error) {
	url := "https://go.dev/dl/?mode=json"
	if includeAll {
		url += "&include=all"
	}
	downloadBaseURL := "https://dl.google.com/go/"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var versions []GolangDlVersion
	err = json.NewDecoder(resp.Body).Decode(&versions)
	if err != nil {
		return nil, err
	}

	var tbs []*Tarball
	for _, v := range versions {
		for _, f := range v.Files {
			if f.Os == GOOS && f.Arch == GOARCH {
				tbs = append(tbs, &Tarball{
					Version: strings.TrimPrefix(f.Version, "go"),
					URL:     downloadBaseURL + f.Filename,
				})
				break
			}
		}
	}

	sort.Sort(tarballSlice(tbs))
	return tbs, nil
}

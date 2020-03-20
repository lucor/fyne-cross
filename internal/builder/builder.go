/*
Package builder implements the build actions for the supperted OS and arch
*/
package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucor/fyne-cross/internal/volume"
)

const (
	baseDockerImage    = "lucor/fyne-cross:develop"
	androidDockerImage = baseDockerImage + "-android"
	linuxDockerImage   = baseDockerImage
	windowsDockerImage = baseDockerImage
	darwinDockerImage  = baseDockerImage

	fyneCmd      = "/usr/local/bin/fyne"
	gowindresCmd = "/usr/local/bin/gowindres"

	defaultIcon = "Icon.png"
)

// Builder represents a builder
type Builder interface {
	PreBuild(vol *volume.Volume, opts PreBuildOptions) error
	Build(vol *volume.Volume, opts BuildOptions) error
	BuildEnv() []string
	BuildLdFlags() []string
	BuildTags() []string
	DockerImage() string
	Package(vol *volume.Volume, opts PackageOptions) error
	Output() string
	TargetID() string
}

// Options specifies the options for creating
// the types that implement the Builder interface
type Options struct {
	Arch        string
	Output      string
	DockerImage string
}

// PreBuildOptions holds the options for the pre build step
type PreBuildOptions struct {
	Verbose bool   // Verbose if true, enable verbose mode
	Icon    string // Icon is the optional icon in png format to use for distribution
	AppID   string // Icon is the appID to use for distribution
}

// BuildOptions holds the options to build the package
type BuildOptions struct {
	Package    string   // Package is the package to build named by the import path as per 'go build'
	LdFlags    []string // LdFlags are the ldflags to pass to the compiler
	Tags       []string // Tags are the tags to pass to the compiler
	StripDebug bool     // StripDebug if true, strips binary output
	Verbose    bool     // Verbose if true, enable verbose mode
}

// PackageOptions holds the options to generate a package for distribution
type PackageOptions struct {
	Icon    string // Icon is the optional icon in png format to use for distribution
	AppID   string // Icon is the appID to use for distribution
	Verbose bool   // Verbose if true, enable verbose mode
}

// goModInit ensure a go.mod exists. If not try to generates a temporary one
func goModInit(b Builder, vol *volume.Volume, verbose bool) error {
	// check if the go.mod exists
	goModPath := filepath.Join(vol.WorkDirHost(), "go.mod")
	_, err := os.Stat(goModPath)
	if err == nil {
		if verbose {
			fmt.Println("go.mod found")
		}
		return nil
	}

	if verbose {
		fmt.Println("go.mod not found, creating a temporary one...")
	}

	// Module does not exists, generate a temporary one
	command := "go mod init fyne-cross-temp-module"
	err = runBuilderDockerCmd(b, vol, []string{}, vol.WorkDirContainer(), []string{command}, verbose)

	if err != nil {
		return fmt.Errorf("Could not generate the temporary go module: %v", err)
	}
	return nil
}

// goBuildCmd returns the go build command
func goBuildCmd(output string, opts BuildOptions) []string {
	// add go build command
	args := []string{"go", "build"}

	ldflags := opts.LdFlags
	// Strip debug information
	if opts.StripDebug {
		ldflags = append(ldflags, "-w", "-s")
	}

	// add ldflags to command, if any
	if len(ldflags) > 0 {
		args = append(args, "-ldflags", fmt.Sprintf("'%s'", strings.Join(ldflags, " ")))
	}

	// add tags to command, if any
	tags := opts.Tags
	if len(tags) > 0 {
		args = append(args, "-tags", fmt.Sprintf("'%s'", strings.Join(tags, " ")))
	}

	args = append(args, "-o", output)

	// add verbose flag
	if opts.Verbose {
		args = append(args, "-v")
	}

	//add package
	args = append(args, opts.Package)
	return args
}

// cp is copies a resource from src to dest
func cp(src string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}

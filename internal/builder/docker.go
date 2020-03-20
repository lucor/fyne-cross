package builder

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/lucor/fyne-cross/internal/volume"
)

type dockerImageProvider interface {
	DockerImageName() (string, error)
}

// Docker image referenced by name. Might be local
// or remote.
type dockerImageName struct {
	Name string
}

func (d *dockerImageName) DockerImageName() (string, error) {
	return d.Name, nil
}

type dockerFile struct {
	Path string
}

func (d *dockerFile) DockerImageName() (string, error) {
	// Hash the Dockerfile to produce an image name
	data, err := ioutil.ReadFile(d.Path)
	if err != nil {
		return "", err
	}
	h := sha1.Sum(data)
	image := hex.EncodeToString(h[:])
	dir := filepath.Dir(d.Path)
	cmd := exec.Command("docker", "build", dir, "-t", image)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return image, nil
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func newDockerImageProvider(imageName string) dockerImageProvider {
	if filepath.Base(imageName) == "Dockerfile" && isFile(imageName) {
		return &dockerFile{Path: imageName}
	}
	// Assume it's an image name
	return &dockerImageName{Name: imageName}
}

// dockerCmd exec a command inside the container for the specified image
func dockerCmd(image string, vol *volume.Volume, env []string, workDir string, command []string, verbose bool) *exec.Cmd {
	// define workdir
	w := vol.WorkDirContainer()
	if workDir != "" {
		w = workDir
	}

	args := []string{
		"run", "--rm", "-t",
		"-w", w, // set workdir
		"-v", fmt.Sprintf("%s:%s", vol.WorkDirHost(), vol.WorkDirContainer()), // mount the working dir
		"-v", fmt.Sprintf("%s:%s", vol.CacheDirHost(), vol.CacheDirContainer()), // mount the cache dir
		"-e", "CGO_ENABLED=1", // enable CGO
		"-e", fmt.Sprintf("GOCACHE=%s", vol.GoCacheDirContainer()), // mount GOCACHE to reuse cache between builds
	}

	// add custom env variables
	for _, e := range env {
		args = append(args, "-e", e)
	}

	// attempt to set fyne user id as current user id to handle mount permissions
	// on linux and MacOS
	if runtime.GOOS != "windows" {
		u, err := user.Current()
		if err == nil {
			args = append(args, "-e", fmt.Sprintf("fyne_uid=%s", u.Uid))
		}
	}

	// specify the image to use
	args = append(args, image)

	// add the command to execute
	args = append(args, command...)

	// run the command inside the container
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if verbose {
		fmt.Println(cmd.String())
	}

	return cmd
}

// runBuilderDockerCmd runs the given command in the docker image returned by the Builder
func runBuilderDockerCmd(builder Builder, vol *volume.Volume, env []string, workDir string, command []string, verbose bool) error {
	// retrieve image
	imageName := builder.DockerImage()
	provider := newDockerImageProvider(imageName)
	image, err := provider.DockerImageName()
	if err != nil {
		return err
	}
	return dockerCmd(image, vol, env, workDir, command, verbose).Run()
}

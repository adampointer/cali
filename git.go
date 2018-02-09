package cali

import (
	"crypto/md5"
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// GitCheckoutConfig is input for Git.Checkout
type GitCheckoutConfig struct {
	Repo, Branch, RelPath, Image string
}

const gitImage = "indiehosters/git:latest"

// Git returns a new instance
func (c *DockerClient) Git() *Git {
	return &Git{c: c, Image: gitImage}
}

// Git is used to interact with containerised git
type Git struct {
	c     *DockerClient
	Image string
}

// GitCheckout will create and start a container, checkout repo and leave container stopped
// so volume can be imported
func (g *Git) Checkout(cfg *GitCheckoutConfig) (string, error) {
	// TODO: this should be more user-friendly!
	// Should include the binary name (e.g. cali), and repo name in plaintext. Maybe branch too.
	// But should also definitely needs an MD5

	repoName, err := repoNameFromUrl(cfg.Repo)
	if err != nil {
		return "", fmt.Errorf("Failed to create data container for %s: %s", cfg.Repo, err)
	}

	// TODO: this should be a separate function
	name := fmt.Sprintf("data_%s_%s_%s_%x",
		repoName,
		strings.Replace(cfg.RelPath, "/", "-", -1),
		strings.Replace(cfg.Branch, "/", "-", -1),
		md5.Sum([]byte(cfg.Repo)),
	)

	if g.c.ContainerExists(name) {
		log.Infof("Existing data container found: %s", name)

		if _, err := g.Pull(name); err != nil {
			log.Warnf("Git pull error: %s", err)
			return name, err
		}
		return name, nil
	} else {
		log.WithFields(log.Fields{
			"git_url": cfg.Repo,
			"image":   g.Image,
		}).Info("Creating data containers")

		co := container.Config{
			Cmd:          []string{"clone", cfg.Repo, "-b", cfg.Branch, "--depth", "1", "."},
			Image:        gitImage,
			Tty:          true,
			AttachStdout: true,
			AttachStderr: true,
			WorkingDir:   "/tmp/workspace",
			Entrypoint:   []string{"git"},
		}
		hc := container.HostConfig{
			Binds: []string{
				"/tmp/workspace",
				fmt.Sprintf("%s/.ssh:/root/.ssh", os.Getenv("HOME")),
			},
		}
		nc := network.NetworkingConfig{}

		g.c.SetConf(&co)
		g.c.SetHostConf(&hc)
		g.c.SetNetConf(&nc)

		id, err := g.c.StartContainer(false, name)

		if err != nil {
			return "", fmt.Errorf("Failed to create data container for %s: %s", cfg.Repo, err)
		}
		return id, nil
	}
}

// repoNameFromUrl takes a git repo URL and returns a string
// representing the repository name
// TODO: tests for this
func repoNameFromUrl(url string) (string, error) {

	// TODO later:
	// remove .git from end
	// remove .*@ at beginning
	// remove protocol

	// Regex for container names: [a-zA-Z0-9][a-zA-Z0-9_.-]
	// but to simplify, just use [a-zA-Z0-9]
	// This regex matches every other character
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return "", fmt.Errorf("Unable to generate repo name: %s", err)
	}
	repoName := reg.ReplaceAllString(url, "-")

	return repoName, nil
}

func (g *Git) Pull(name string) (string, error) {
	co := container.Config{
		Cmd:          []string{"pull"},
		Image:        g.Image,
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/tmp/workspace",
		Entrypoint:   []string{"git"},
	}
	hc := container.HostConfig{
		VolumesFrom: []string{name},
		Binds: []string{
			fmt.Sprintf("%s/.ssh:/root/.ssh", os.Getenv("HOME")),
		},
	}
	nc := network.NetworkingConfig{}

	g.c.SetConf(&co)
	g.c.SetHostConf(&hc)
	g.c.SetNetConf(&nc)

	return g.c.StartContainer(true, "")
}

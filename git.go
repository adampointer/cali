package main

import (
	"crypto/md5"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// GitCheckoutConfig is input for Git.Checkout
type GitCheckoutConfig struct {
	Repo, Branch, RelPath string
}

const gitImage = "hub.platformservices.io/platformservices/git:1.1"

// Git returns a new instance
func (c *DockerClient) Git() *Git {
	return &Git{c: c}
}

// Git is used to interact with containerised git
type Git struct {
	c *DockerClient
}

// GitCheckout will create and start a container, checkout repo and leave container stopped
// so volume can be imported
func (g *Git) Checkout(cfg *GitCheckoutConfig) (string, error) {

	if cfg.RelPath == "" {
		cfg.RelPath = "."
	}
	name := fmt.Sprintf("data_%s_%s_%x", cfg.RelPath, cfg.Branch, md5.Sum([]byte(cfg.Repo)))

	if g.c.ContainerExists(name) {
		log.Infof("Existing data container found: %s", name)

		if _, err := g.Pull(name, cfg.RelPath); err != nil {
			log.Warnf("Git pull error: %s", err)
			return name, err
		}
		return name, nil
	} else {
		log.WithFields(log.Fields{
			"git_url": cfg.Repo,
			"image":   gitImage,
		}).Info("Creating data containers")

		co := container.Config{
			Cmd:          []string{"clone", cfg.Repo, "-b", cfg.Branch, "--depth", "1", cfg.RelPath},
			Image:        gitImage,
			Tty:          true,
			AttachStdout: true,
			AttachStderr: true,
		}
		hc := container.HostConfig{}
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

func (g *Git) Pull(name, relPath string) (string, error) {
	co := container.Config{
		Cmd:          []string{"pull"},
		Image:        gitImage,
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   fmt.Sprintf("/tmp/workspace/%s", relPath),
	}
	hc := container.HostConfig{
		VolumesFrom: []string{name},
	}
	nc := network.NetworkingConfig{}

	g.c.SetConf(&co)
	g.c.SetHostConf(&hc)
	g.c.SetNetConf(&nc)

	return g.c.StartContainer(true, "")
}

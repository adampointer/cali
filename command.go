package main

import (
	"log"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

type FlagSetter func(flags *flag.FlagSet)

type InitFunc func(args []string)

type AppConfig struct {
	Image, WorkDir   string
	Cmd, Envs, Binds []string
	Privileged       bool
}

type DockerTask interface {
	Init(AppConfig, *DockerClient, []string)
}

type Task struct{}

func (t *Task) Init(cfg AppConfig, cli *DockerClient, args []string) {
	cli.SetImage(cfg.Image)
	cli.SetCmd(cfg.Cmd)
	cli.SetBinds(cfg.Binds)
	cli.SetEnvs(cfg.Envs)
	cli.SetWorkDir(cfg.WorkDir)
	cli.Privileged(cfg.Privileged)
}

type Command struct {
	Name, ShortDescription, LongDescription string
	Init                                    InitFunc
	RunCfg                                  AppConfig
	Task                                    DockerTask
	Flags                                   FlagSetter
	cmd                                     *cobra.Command
}

func (c *Command) generate() {
	c.cmd = &cobra.Command{
		Use:   c.Name,
		Short: c.ShortDescription,
		Long:  c.LongDescription,
		PreRun: func(cmd *cobra.Command, args []string) {
			if c.Init != nil {
				c.Init(args)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var task DockerTask

			if c.Task == nil {
				task = new(Task)
			} else {
				task = c.Task
			}

			cli := GetDockerClient()
			task.Init(c.RunCfg, cli, args)
			if _, err := cli.StartContainer(true, ""); err != nil {
				log.Fatalf("Error executing task: %s", err)
			}
		},
	}
	if c.Flags != nil {
		c.Flags(c.cmd.PersistentFlags())
	}
}

type CommandIter func(name string, cmd *Command) error

type Commands map[string]*Command

func (c Commands) forEach(f CommandIter) error {
	for k, v := range c {
		if err := f(k, v); err != nil {
			return err
		}
	}
	return nil
}

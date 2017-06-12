package main

import (
	"fmt"
	"os"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	debug, jsonLogs, nonInteractive bool
	dockerHost, cfgFile             string
)

type Application struct {
	Name, ShortDescription, LongDescription string
	commands                                Commands
	cmd                                     *cobra.Command
}

func (a *Application) RunWith(c Commands) {
	c.forEach(func(name string, cmd *Command) error {
		a.add(name, cmd)
		return nil
	})
	if err := a.initCobra(); err != nil {
		os.Exit(1)
	}
}

func (a *Application) add(name string, cmd *Command) {
	cmd.Name = name
	if a.commands == nil {
		a.commands = make(Commands)
	}
	cmd.generate()
	a.commands[name] = cmd
}

func (a *Application) generate() {
	a.cmd = &cobra.Command{
		Use:   a.Name,
		Short: a.ShortDescription,
		Long:  a.LongDescription,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Setup logging
			if debug {
				log.SetLevel(log.DebugLevel)
			}

			if jsonLogs {
				log.SetFormatter(&log.JSONFormatter{})
			}
		},
	}
}

func (a *Application) initCobra() error {
	a.generate()
	a.commands.forEach(func(name string, c *Command) error {
		a.cmd.AddCommand(c.cmd)
		return nil
	})
	if err := a.cmd.Execute(); err != nil {
		return err
	}
	return nil
}

type InitFunc func(args []string)

type AppConfig struct {
	Image, WorkDir   string
	Cmd, Envs, Binds []string
	Privileged       bool
}

type DockerTask interface {
	Start(AppConfig, []string) error
}

type Task struct{}

func (t *Task) Start(cfg AppConfig, args []string) error {
	cli := GetDockerClient()
	cli.SetImage(cfg.Image)
	cli.SetCmd(cfg.Cmd)
	cli.SetBinds(cfg.Binds)
	cli.SetEnvs(cfg.Envs)
	cli.SetWorkDir(cfg.WorkDir)
	cli.Privileged(cfg.Privileged)
	_, err := cli.StartContainer(true, "")
	return err
}

type Command struct {
	Name, ShortDescription, LongDescription string
	Init                                    InitFunc
	RunCfg                                  AppConfig
	Task                                    DockerTask
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
			if err := task.Start(c.RunCfg, args); err != nil {
				log.Fatalf("Error executing task: %s", err)
			}
		},
	}
	initRootCmd(c.cmd)
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

// Init attaches to the parent command
func initRootCmd(rootCmd *cobra.Command) {
	cobra.OnInitialize(initConfig)

	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pscli.yaml)")

	// this should be done with build tags, but for the sake of expediency we'll check it at runtime
	var dockerSocket string
	if runtime.GOOS == "windows" {
		dockerSocket = "npipe:////./pipe/docker_engine"
	} else {
		dockerSocket = "unix:///var/run/docker.sock"
	}
	rootCmd.PersistentFlags().StringVarP(&dockerHost, "docker-host", "H", dockerSocket, "URI of Docker Daemon")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
	rootCmd.PersistentFlags().BoolVarP(&jsonLogs, "json", "j", false, "Log in json format")
	rootCmd.PersistentFlags().BoolVarP(&nonInteractive, "non-interactive", "N", false, "Do not use a tty, non-interactive output only")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".pscli") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AutomaticEnv()          // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

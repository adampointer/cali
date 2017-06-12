package main

import (
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func mooFlags(f *flag.FlagSet) {
	var utterance string
	f.StringVarP(&utterance, "utterance", "u", "Hello", "What should the cow say?")
	viper.BindPFlag("utterance", f.Lookup("utterance"))
}

type mooTask struct{}

func (m *mooTask) Init(cfg AppConfig, cli *DockerClient, args []string) {
	cli.SetImage(cfg.Image)
	cli.SetCmd([]string{"/usr/games/cowsay", viper.GetString("utterance")})
}

func main() {
	app := Application{
		Name:             "test",
		ShortDescription: "Test application",
		LongDescription:  "An application to test with",
	}
	app.RunWith(Commands{
		"moo": {
			ShortDescription: "Cowsay",
			LongDescription:  "The cow, he say moo",
			RunCfg: AppConfig{
				Image: "chuanwen/cowsay:latest",
			},
			Flags: mooFlags,
			Task:  new(mooTask),
		},
	})
}

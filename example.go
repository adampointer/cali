package main

func main() {
	cli := Cli("examplecli")
	cli.SetShort("Example CLI tool")
	cli.SetLong("A nice long description of what your tool actually does")

	moo := cli.Command("moo")
	moo.SetShort("Cow says moo")
	moo.SetLong("What does the cow say?")
	moo.Flags().StringP("utterance", "u", "Hello", "What should the cow say?")
	moo.BindFlags()

	mooTask := moo.Task("chuanwen/cowsay:latest")
	mooTask.SetInitFunc(func(t *Task, args []string) {
		t.SetCmd([]string{"/usr/games/cowsay", cli.FlagValues().GetString("utterance")})
	})

	cli.Start()
}

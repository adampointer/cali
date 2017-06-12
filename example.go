package main

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
		},
	})
}

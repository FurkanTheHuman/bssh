package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"example.com/bssh/config"
	ssh "example.com/bssh/ssh_common"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"github.com/urfave/cli/v2"
)

// const configPath string = "./config"

var DEBUG = os.Getenv("BSSH_DEBUG")

func main() {
	namespaceFlag := []cli.Flag{
		&cli.StringFlag{
			Name:    "namespace",
			Aliases: []string{"n"},
			Usage:   "filter by namespace",
		},
	}

	app := &cli.App{
		Name:  "bssh",
		Usage: "Timeless ssh manager",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "namespace",
				Aliases: []string{"n"},
				Usage:   "specify namespace",
			},
			&cli.BoolFlag{
				Name:    "show-password",
				Aliases: []string{"s"},
				Usage:   "specify password",
			},
		},

		Action: func(c *cli.Context) error {

			sshs, err := config.GetSshList()
			var namespace string

			if c.Bool("namespace") {

				namespaces := config.GetNamespaces(sshs)
				selected, _ := fuzzyfinder.Find(namespaces, func(i int) string {
					return namespaces[i]
				}, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
					if i == -1 {
						return ""
					}

					var meta string
					filtered_list, _ := config.FilterSsh(sshs, namespaces[i])
					for filtered, _ := range filtered_list {
						meta = meta + sshs[filtered].Username + "@" + sshs[filtered].Addr + "\n"

					}
					return fmt.Sprintf(meta)
				}))
				namespace = namespaces[selected]
			}
			if namespace != "" {
				sshs, _ = config.FilterSsh(sshs, namespace) // Dont ignore this err

			}
			if err != nil {
				fmt.Println("failed to run fzf!")

				if DEBUG == "true" {
					log.Println(err)
				}
				return err
			}
			idx, err := fuzzyfinder.Find(sshs, func(i int) string {
				return sshs[i].Username + "@" + sshs[i].Addr
			},
				fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
					if i == -1 {
						return ""
					}

					checker := func(s config.SshSource) string {
						if s.Password == "" {
							return "<USES PRIVATE KEY>"
						}
						return s.Password
					}

					return fmt.Sprintf("ssh: %s (%s) \nPassword: %s \nNamespace: %s",
						sshs[i].Username,
						sshs[i].Addr,
						checker(sshs[i]),
						sshs[i].Namespace)
				}))
			if err != nil {
				os.Exit(1)
			}
			ssh.Connect(sshs[idx])

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "add", //add in future,
				Aliases: []string{"a", "ssh"},
				Usage:   "add a task to the list",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "namespace",
						Aliases: []string{"n"},
						Value:   "default",
						Usage:   "specify namespace",
					},
					&cli.StringFlag{
						Name:    "password",
						Aliases: []string{"p"},
						Usage:   "specify password",
					},
					&cli.StringFlag{
						Name:    "key",
						Aliases: []string{"k", "i"},
						Usage:   "specify ssh key file path",
					},
					&cli.StringFlag{
						Name:     "addr",
						Aliases:  []string{"a"},
						Usage:    "specify ssh ssh address to connect (username@ip_or_dns)",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {

					user, addr := config.SplitFullname(c.String("addr"))
					port := "22"
					password := c.String("password")
					var key []byte
					if c.String("key") != "" {
						pemBytes, err := ioutil.ReadFile(c.String("key"))
						key = pemBytes
						if err != nil {
							log.Fatalln("Could not read private key file, Ensure you have the right permissions.")
						}

					}
					s := config.SshSource{Username: user, Addr: addr, Password: password, Port: port, Key: key, Namespace: c.String("namespace")}
					b, err := json.MarshalIndent(s, "", "\t")
					if err != nil {
						log.Fatalln("Can not parse SshSource")
					}
					config.UpdateConfigFile(string(b))
					return nil
				},
			},
			{
				Name:    "list", //add in future,
				Aliases: []string{"ls"},
				Flags:   namespaceFlag,
				Usage:   "list available ssh connections",
				Action: func(c *cli.Context) error {
					x := c.String("namespace")
					fmt.Println(x)
					contents, err := config.GetSshList()
					if err != nil {
						log.Fatalln("FAILED", err)
					}
					for i, s := range contents {
						fmt.Printf("[%d] %s@%s\n", i, s.Username, s.Addr)
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

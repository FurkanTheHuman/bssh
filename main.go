package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	b64 "encoding/base64"

	"example.com/bssh/bucket"
	"example.com/bssh/fzf"
	ssh "example.com/bssh/ssh_common"
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
				Usage:   "filter by namespace",
			},
			&cli.BoolFlag{
				Name:    "show-password",
				Aliases: []string{"s"},
				Usage:   "show password",
			},
		},

		Action: func(c *cli.Context) error {
			if c.Args().Len() > 0 {
				cli.ShowAppHelp(c)
				fmt.Println("\n" + c.Args().Get(0) + " is not a command")
				os.Exit(1)
			}
			selectedSsh, err := fzf.FuzzySshSelector(c.Bool("namespace"))

			if err != nil {
				os.Exit(1)
			}

			return ssh.Connect(selectedSsh)
		},
		Commands: []*cli.Command{
			{
				Name:    "add", //port flag needed,
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
					&cli.StringFlag{
						Name:     "alias",
						Usage:    "create an alias for better search",
						Required: false,
					},
				},
				Action: func(c *cli.Context) error {

					user, addr := bucket.SplitFullname(c.String("addr"))
					port := "22"
					password := c.String("password")
					password = b64.StdEncoding.EncodeToString([]byte(password))
					alias := c.String("alias")
					var key []byte
					if c.String("key") != "" {
						pemBytes, err := ioutil.ReadFile(c.String("key"))
						key = pemBytes
						if err != nil {
							log.Fatalln("Could not read private key file, Ensure you have the right permissions.")
						}

					}
					s := bucket.SshSource{Username: user, Addr: addr, Password: password, Port: port, Key: key, Namespace: c.String("namespace"), Alias: alias}

					bucket.UpdateConfigFile(s)
					return nil
				},
			},
			{
				Name:    "list", //Maybe categorize the output and make it look nice,
				Aliases: []string{"ls"},
				Flags:   namespaceFlag,
				Usage:   "list available ssh connections",
				Action: func(c *cli.Context) error {
					namespace := c.String("namespace")
					contents, err := bucket.GetSshList()
					if err != nil {
						log.Fatalln("FAILED", err)
					}
					if namespace != "" {
						contents = bucket.FilterSsh(contents, namespace)
					}
					if len(contents) <= 1 {
						fmt.Println("This namespace is empty")
					}
					for i, s := range contents {
						fmt.Printf("[%d] - Namespace: %s,\t Address: %s@%s\n", i, s.Namespace, s.Username, s.Addr)
					}
					return nil
				},
			},
			{
				Name:    "remove", //add in future,
				Aliases: []string{"r", "rm"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "namespace",
						Aliases: []string{"n"},
						Usage:   "filter by namespace",
					},
				},
				Usage: "remove selected connection from bucket",
				Action: func(c *cli.Context) error {
					s, err := fzf.FuzzySshSelector(c.Bool("namespace"))
					if err != nil {
						os.Exit(0)
					}
					_, err = bucket.RemoveSsh(s)
					return err
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

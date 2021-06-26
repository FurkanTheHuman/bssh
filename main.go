package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	b64 "encoding/base64"

	"github.com/furkanthehuman/bssh/bucket"
	"github.com/furkanthehuman/bssh/fzf"
	ssh "github.com/furkanthehuman/bssh/ssh_common"
	"github.com/jedib0t/go-pretty/v6/table"
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
		Usage: "Simple ssh manager",
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
					tw := table.NewWriter()
					tw.AppendHeader(table.Row{"#", "Ssh Address", "Namespace", "Uses", "Alias"})

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
						if s.Password != "" {

							tw.AppendRow(table.Row{i + 1, s.Username + "@" + s.Addr, s.Namespace, "password", s.Alias})
						} else {
							tw.AppendRow(table.Row{i + 1, s.Username + "@" + s.Addr, s.Namespace, "ssh key", s.Alias})

						}
					}

					tw.AppendFooter(table.Row{"", "Total:", len(contents)})
					fmt.Println(tw.Render())

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
					extracted, err := bucket.RemoveSsh(s)
					if extracted.Addr == "" {
						fmt.Println("Entry not found")
						return err
					}
					fmt.Printf("connection %s removed", extracted.Addr)
					return err
				},
			},
			{
				Name:    "ping",
				Aliases: []string{"p"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "namespace",
						Aliases: []string{"n"},
						Usage:   "ping all of the namespace",
					},
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "ping all entries",
					},
				},
				Usage: "ping servers for availability",
				Action: func(c *cli.Context) error {
					ssh_list, err := bucket.GetSshList()
					if c.Bool("namespace") {
						s, err := fzf.FuzzyNamespaceSelector(ssh_list)
						if err != nil {
							os.Exit(1)
						}
						ssh.PingNamespace(s)
					} else if c.Bool("all") {
						var wg sync.WaitGroup
						sshList, _ := bucket.GetSshList()
						for _, s := range sshList {
							wg.Add(1)
							go ssh.PingWorker(s, &wg)
						}
						wg.Wait()
					} else {
						s, err := fzf.FuzzySshSelector(false)
						if err != nil {
							os.Exit(1)
						}
						ssh.Ping(s)
					}

					return err
				},
			},
			{
				Name:    "run",
				Aliases: []string{"x"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "namespace",
						Aliases: []string{"n"},
						Usage:   "ping all of the namespace",
					},
				},
				Usage: "run commands on remote",
				Action: func(c *cli.Context) error {
					ssh_list, err := bucket.GetSshList()
					if c.Bool("namespace") {
						ns, err := fzf.FuzzyNamespaceSelector(ssh_list)
						if err != nil {
							os.Exit(1)
						}
						ssh.SendCommandToNamespace(ns, strings.Join(c.Args().Slice()[:], " "))

					} else {
						s, err := fzf.FuzzySshSelector(false)
						if err != nil {
							os.Exit(1)
						}
						output, err := ssh.SendCommand(s, strings.Join(c.Args().Slice()[:], " "))
						if err != nil {
							log.Fatal(err)
						}
						fmt.Println(string(output))

					}

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

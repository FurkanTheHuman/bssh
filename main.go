package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
)

func main() {
	args := os.Args

	config := &ssh.ClientConfig{
		User: args[2],
		Auth: []ssh.AuthMethod{
			ssh.Password(args[3]),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := DialHost(args[1], config)

	if err != nil {
		os.Exit(1)
	}

	session, err := client.NewSession()

	fd, origState := RequestTerminal(session)
	err = session.Shell()
	if err != nil {
		log.Fatal("LINE: X -> Can't  save original terminal state ")

	}
	defer terminal.Restore(fd, origState)
	session.Wait()
	fmt.Println("Exitting goodbye")

}

func DialHost(addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Printf(" Dialing failed for %s ", addr)
		log.Printf(" ERR: ", err)

	}
	return client, err
}

func RequestTerminal(session *ssh.Session) (int, *term.State) {
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	fileDescriptor := int(os.Stdin.Fd())

	if terminal.IsTerminal(fileDescriptor) {
		originalState, err := terminal.MakeRaw(fileDescriptor)
		if err != nil {
			log.Fatal("LINE: X -> Can't  save original terminal state ")
		}

		termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
		if err != nil {
			log.Fatal("LINE: X -> Can't  save original terminal state ")

		}

		err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if err != nil {
			log.Fatal("LINE: X -> Can't  save original terminal state ")

		}
		return fileDescriptor, originalState
	}
	return -1, nil
}

func WriteFile(path string, newContent string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	f.WriteString(newContent)
}

package ssh

import (
	"fmt"
	"log"
	"os"

	b64 "encoding/base64"

	bucket "github.com/FurkanTheHuman/bssh/bucket"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
)

func Connect(s bucket.SshSource) error {
	var config *ssh.ClientConfig

	if s.Password != "" {
		decodedPassword, err := b64.StdEncoding.DecodeString(s.Password)
		if err != nil {
			fmt.Println("ERROR: password is corrupted for this host. Can not decode.")
		}
		config = &ssh.ClientConfig{
			User: s.Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(string(decodedPassword)),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else {
		signer, err := ssh.ParsePrivateKey(s.Key)
		if err != nil {

			if _, ok := err.(*ssh.PassphraseMissingError); ok {
				var passphrase string
				fmt.Scanln(&passphrase)

				signer, err = ssh.ParsePrivateKeyWithPassphrase(s.Key, []byte(passphrase))
				if err != nil {
					log.Fatalln("Passphrase rejected")
				}
			} else {
				log.Fatalln("can not parse Private Key")
			}

		}
		config = &ssh.ClientConfig{
			User: s.Username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}

	client, err := DialHost(s.Addr+":"+s.Port, config)

	if err != nil {
		os.Exit(1)
	}

	session, err := client.NewSession()

	if err != nil {
		log.Fatal("Could not create a new session...")
	}
	fd, origState := RequestTerminal(session)
	err = session.Shell()
	if err != nil {
		log.Fatal("LINE: X -> Can't  save original terminal state ")

	}
	defer terminal.Restore(fd, origState)
	return session.Wait()
}

func DialHost(addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Printf(" Dialing failed for %s ", addr)
		log.Printf(" ERR: %s", err)

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
			log.Fatal("Can't  save original terminal state ")
		}

		termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
		if err != nil {
			log.Fatal("Can't  save original terminal state ")

		}

		err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if err != nil {
			log.Fatal("Can't  save original terminal state ")

		}
		return fileDescriptor, originalState
	}
	return -1, nil
}

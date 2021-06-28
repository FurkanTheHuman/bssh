package ssh

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	b64 "encoding/base64"

	"github.com/fatih/color"
	bucket "github.com/furkanthehuman/bssh/bucket"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
)

func Connect(s bucket.SshSource) error {
	config := GetSshConfig(s)

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

func GetSshConfig(s bucket.SshSource) *ssh.ClientConfig {
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
			Timeout:         time.Second * 5,
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
			Timeout:         time.Second * 5,
		}
	}
	return config
}

func DialHost(addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	client, err := ssh.Dial("tcp", addr, config)

	return client, err
}

func PingWorker(s bucket.SshSource, wg *sync.WaitGroup) {
	defer wg.Done()
	// colors for messages
	Ping(s)
}

func Ping(s bucket.SshSource) {
	// colors for messages
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	_, err := DialHost(s.Addr+":"+s.Port, GetSshConfig(s))
	alias := ""
	if s.Alias != "" {
		alias = "(" + s.Alias + ")"
	}
	if err != nil {

		if strings.Contains(err.Error(), "timeout") {
			fmt.Printf("%s %s@%s %s is timed out. \n", cyan("[?]"), s.Username, s.Addr, alias)

		} else {
			fmt.Printf("%s %s@%s %s is not accesible\n", red("[X]"), s.Username, s.Addr, alias)
		}
	} else {
		fmt.Printf("%s %s@%s %s is OK\n", green("[✔]"), s.Username, s.Addr, alias)
	}
}

func PingNamespace(namespace string) {
	sshList, _ := bucket.GetSshList()
	filteredList := bucket.FilterSsh(sshList, namespace)
	//const numJobs = 8
	//jobs := make(chan bucket.SshSource, numJobs)
	// results := make(chan int, numJobs)
	var wg sync.WaitGroup
	for _, s := range filteredList {
		wg.Add(1)
		go PingWorker(s, &wg)
	}
	wg.Wait()

}

func SendCommand(s bucket.SshSource, commands ...string) ([]byte, error) {
	config := GetSshConfig(s)

	client, err := DialHost(s.Addr+":"+s.Port, config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()

	if err != nil {
		return nil, err
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		return []byte{}, err
	}

	in, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}

	out, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	var output []byte
	wait := make(chan bool)

	go func(in io.WriteCloser, out io.Reader, output *[]byte, done chan bool) {
		var (
			line string
			r    = bufio.NewReader(out)
		)
		for {

			b, err := r.ReadByte()
			if err != nil {
				break
			}

			*output = append(*output, b)

			if b == byte('\n') {
				line = ""
				continue
			}

			line += string(b)

			if strings.HasPrefix(line, "[sudo] password for ") && strings.HasSuffix(line, ": ") {
				if s.Password != "" {
					_, err = in.Write([]byte(s.Password + "\n"))
					if err != nil {

						break
					}
				} else {
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("This command requires sudo password: ")
					password, _ := reader.ReadString('\n')
					_, err = in.Write([]byte(password + "\n"))
					if err != nil {

						break
					}
				}

			}

		}
		done <- true
	}(in, out, &output, wait)

	cmd := strings.Join(commands, "; ")

	_, err = session.Output(cmd)
	<-wait
	time.Sleep(time.Second)
	if err != nil {
		return []byte{}, err
	}

	return output, nil
}

func SendCommandWorker(s bucket.SshSource, wg *sync.WaitGroup, mutex *sync.Mutex, commands ...string) {
	defer wg.Done()
	output, err := SendCommand(s, commands...)
	black := color.New(color.FgWhite)
	red := color.New(color.FgRed)
	bold := black.Add(color.Bold)
	boldRed := red.Add(color.Bold)
	mutex.Lock()
	if err != nil {
		boldRed.Println("Error on : " + s.Username + "@" + s.Addr + " (" + s.Alias + "):")
		boldRed.Println(err)

	} else {
		color.Green("===============")
		bold.Println("START OF: " + s.Username + "@" + s.Addr + " (" + s.Alias + "):\n")
		fmt.Println(string(output))
		bold.Println("END OF: " + s.Username + "@" + s.Addr + " (" + s.Alias + "):")
		color.Green("===============\n")

	}
	mutex.Unlock()

}

func SendCommandToNamespace(namespace string, commands ...string) {
	sshList, _ := bucket.GetSshList()
	filteredList := bucket.FilterSsh(sshList, namespace)
	var mutex = &sync.Mutex{}

	var wg sync.WaitGroup
	for _, s := range filteredList {
		wg.Add(1)
		go SendCommandWorker(s, &wg, mutex, commands...)
	}
	wg.Wait()
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

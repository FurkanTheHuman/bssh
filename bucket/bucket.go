package bucket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
)

const configPath string = "./sshconfig"

type SshSource struct {
	Username  string `json:"hostname"`
	Addr      string `json:"addr"`
	Key       []byte `json:"key"`
	Password  string `json:"password"`
	Port      string `json:"port"`
	Namespace string `json:"namespace"`
	Alias     string `json:"alias"`
	// maybe add date here
}

func (s SshSource) CompileCommand() string {
	var result string
	fullAddr := s.CompileFullAddr()
	isPort22, port := s.GetPort()
	PasswordAsComment := s.GetPassword()
	if isPort22 { // maybe this shoul be something more generic so I can programaticaly generate it, maybe not
		result = "ssh " + fullAddr + " " + PasswordAsComment

	} else {
		result = "ssh " + fullAddr + " " + port + " " + PasswordAsComment

	}
	return result
}

func (s SshSource) CompileFullAddr() string {
	return s.Username + "@" + s.Addr
}

func (s SshSource) GetPort() (bool, string) {
	if s.Port == "22" {
		return true, ""
	}
	return false, "-p " + s.Port

}

func (s SshSource) GetPassword() string {
	return "# password: " + s.Password
}

func WriteFile(path string, newContent string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	f.WriteString(newContent)
}

func GetFileContents() (string, error) {
	var file []byte
	path := configPath
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("Config file does not exist. Creating one...")
		_, err := os.Create(path)
		f, _ := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
		f.WriteString("[]")
		if err != nil {
			log.Println("Can not create one. Exiting")
			return "", err

		}

		return "", nil
	}
	return string(file), nil
}

func RemoveSsh(id SshSource) (SshSource, error) {
	list, err := GetSshList()
	if err != nil {
		os.Exit(0)
	}

	var filtered []SshSource
	var extracted SshSource
	seen := true // to save the duplicates
	for i := range list {
		if reflect.DeepEqual(list[i], id) && seen {
			extracted = list[i]
			seen = false
			continue
		}
		filtered = append(filtered, list[i])
	}

	b, err := json.MarshalIndent(filtered, "", "\t")
	if err != nil {
		log.Fatalln("config file is corrupted!")
	}
	// there should be eero checking
	f, _ := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if _, err = f.WriteString(string(b)); err != nil {
		log.Println("Write to file failed")

		panic(err)
	}

	return extracted, err
}

func GetSshList() ([]SshSource, error) {
	contents, err := GetFileContents()
	if err != nil {
		return nil, err
	}
	var list []SshSource
	err = json.Unmarshal([]byte(contents), &list)
	if err != nil {
		log.Println("Could not parse config file")
		return nil, err
	}
	return list, nil
}

func FilterSsh(list []SshSource, namespace string) []SshSource {

	var filtered []SshSource
	for i := range list {
		if list[i].Namespace == namespace {
			filtered = append(filtered, list[i])
		}
	}
	return filtered

}

func GetNamespaces(s []SshSource) []string {
	var filtered []string
	seen := make(map[string]struct{}, len(s))
	counter := 0
	for i := range s {
		if _, ok := seen[s[i].Namespace]; ok {
			continue
		}
		seen[s[i].Namespace] = struct{}{}
		counter++
		filtered = append(filtered, s[i].Namespace)
	}
	return filtered[:counter]

}

func UpdateConfigFile(s SshSource) error {
	b, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		log.Fatalln("Can not parse SshSource")
	}
	newContent := string(b)
	path := configPath
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println("Could not read file")
		panic(err)
	}

	list := make([]SshSource, 0)
	var new SshSource
	old, _ := GetFileContents()
	err = json.Unmarshal([]byte(old), &list)
	if err != nil {
		fmt.Println("Config file is empty. Initialising...")

	}

	err = json.Unmarshal([]byte(newContent), &new)
	if err != nil {
		log.Println("Unmarshal failed")

		panic(err)
	}
	list = append(list, new)
	log.Println("There are ", len(list), "records in config")
	b, err = json.MarshalIndent(list, "", "  ")
	if err != nil {
		log.Println("Could not create json from list")

		panic(err)
	}
	if _, err = f.WriteString(string(b)); err != nil {
		log.Println("Write to file failed")

		panic(err)
	}
	return err
}

func SplitFullname(fullname string) (string, string) {
	arr := strings.Split(fullname, "@")
	return arr[0], arr[1]
}

package fzf

import (
	"fmt"
	"log"
	"os"

	config "example.com/bssh/bucket"
	"github.com/ktr0731/go-fuzzyfinder"
)

func FuzzySshSelector(isNamespaced bool) (config.SshSource, error) {
	sshs, err := config.GetSshList()
	if err != nil {
		log.Fatalln("READ FAILED FOR BUCKET")
	}
	var namespace string
	if isNamespaced {
		namespace, err = FuzzyNamespaceSelector(sshs)
		if err != nil {
			os.Exit(0)
		}
		sshs = config.FilterSsh(sshs, namespace)
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
	return sshs[idx], err
}

func FuzzyNamespaceSelector(sshs []config.SshSource) (string, error) {
	sshs, err := config.GetSshList()
	if err != nil {
		log.Fatalln("READ FAILED FOR BUCKET")
	}
	namespaces := config.GetNamespaces(sshs)
	selected, err := fuzzyfinder.Find(namespaces, func(i int) string {
		return namespaces[i]
	}, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		if i == -1 {
			return ""
		}

		var meta string
		filtered_list := config.FilterSsh(sshs, namespaces[i])
		for filtered, _ := range filtered_list {
			meta = meta + filtered_list[filtered].Username + "@" + filtered_list[filtered].Addr + "\n"

		}
		return fmt.Sprintln(meta)
	}))

	return namespaces[selected], err
}

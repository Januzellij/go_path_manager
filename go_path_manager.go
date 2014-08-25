package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"

	"github.com/codegangsta/cli"
)

type Location struct {
	Path string
	Line int
}

func getPathLocation() Location {
	config := map[string][]string{
		"zsh": []string{
			"~/.zshrc",
			"/etc/profile",
			"~/.profile",
			"/etc/zshenv",
			"/etc/zprofile",
			"/etc/zshrc",
			"/etc/zlogin",
			"/etc/zlogout",
			"~/.zshenv",
			"~/.zprofile",
			"~/.zlogin",
		},
		"bash": []string{
			"/etc/profile",
			"~/.profile",
			"~/.bash_profile",
			"~/.bash_login",
			"~/.bash_logout",
			"~/.bashrc",
		},
		"sh": []string{
			"/etc/profile",
			"~/.profile",
		},
		"ksh": []string{
			"/etc/profile",
			"~/.profile",
			"~/.kshrc",
		},
	}

	usercmd := exec.Command("whoami")
	rawUser, err := usercmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	username := strings.Replace(string(rawUser), "\n", "", -1)
	infocmd := exec.Command("finger", username)
	info, err := infocmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	shellRe := regexp.MustCompile(`Shell: \/bin\/([a-zA-Z\/].*)`)
	shellStr := shellRe.FindString(string(info))
	if shellStr == "" {
		log.Fatal("Couldn't find user shell")
	}

	shell := strings.Split(shellStr, "/")[2]
	paths := config[shell]
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	for i, v := range paths {
		paths[i] = strings.Replace(v, "~", usr.HomeDir, -1)
	}

	search := regexp.MustCompile(`^(?:export )?PATH=`)
	location := Location{}
	for _, v := range paths {
		file, err := os.Open(v)
		if err != nil {
			continue
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		i := 0
		for scanner.Scan() {
			if search.Match([]byte(scanner.Text())) {
				location.Path = v
				location.Line = i + 1
				goto done
			}
			i++
		}
	}

	log.Fatal("couldn't find $PATH")

done:
	return location
}

func containsString(list []string, a string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func main() {
	app := cli.NewApp()
	app.Name = "go_path_manager"
	app.Usage = "manage your path variable"
	app.Commands = []cli.Command{
		{
			Name:      "list",
			ShortName: "l",
			Usage:     "Lists the directories in your PATH variable.",
			Action: func(c *cli.Context) {
				fmt.Println("This is your current $PATH:")
				pathItems := strings.Split(os.Getenv("PATH"), ":")
				for i, v := range pathItems {
					if i < 9 {
						fmt.Printf("%d.  %s\n", i+1, v)
					} else {
						fmt.Printf("%d. %s\n", i+1, v)
					}
				}
			},
		},
		{
			Name:      "prepend",
			ShortName: "p",
			Usage:     "Prepends the directory to your PATH variable",
			Action: func(c *cli.Context) {
				fmt.Println("added task: ", c.Args().First())
			},
		},
		{
			Name:      "append",
			ShortName: "a",
			Usage:     "Appends the directory to your PATH variable",
			Action: func(c *cli.Context) {
				fmt.Println("added task: ", c.Args().First())
			},
		},
		{
			Name:      "which",
			ShortName: "w",
			Usage:     "Shows the the different path entry where the program appears. Like which, but not just with the first location",
			Action: func(c *cli.Context) {
				fmt.Println("Warning: If you use rbenv, then the order may be incorrect for gems.")
				program := c.Args().First()
				path := strings.Split(os.Getenv("PATH"), ":")
				var directories []string
				for _, dir := range path {
					file, err := os.Open(dir)
					if err != nil {
						continue
					}
					defer file.Close()
					entries, err := file.Readdirnames(0)
					if err != nil {
						log.Fatal(err)
					}
					if containsString(entries, program) {
						directories = append(directories, dir+"/"+program)
					}
				}
				for i, dir := range directories {
					fmt.Printf("%d. %s\n", i+1, dir)
				}
			},
		},
	}
	app.Run(os.Args)
}

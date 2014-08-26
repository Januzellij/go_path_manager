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
	Path  string
	Index int
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
				location.Index = i
				goto done
			}
			i++
		}
	}

	log.Fatal("couldn't find $PATH")

done:
	return location
}

func addToPath(path string, shouldPrepend bool) {
	location := getPathLocation()
	file, err := os.Open(location.Path)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if shouldPrepend {
		pathLine := lines[location.Index]
		pathItems := strings.Split(pathLine, "=")
		lines[location.Index] = "export PATH=" + path + ":" + pathItems[1]
	} else {
		pathItems := strings.Split(lines[location.Index], ":")
		lastIndice := len(pathItems) - 1
		if pathItems[lastIndice] == "$PATH" {
			pathItems = pathItems[:lastIndice]
			pathItems = append(pathItems, path, "$PATH")
		} else {
			pathItems = append(pathItems, path)
		}
		pathString := strings.Join(pathItems, ":")
		lines[location.Index] = pathString
	}

	file.Close()
	os.Remove(location.Path)
	newFile, err := os.Create(location.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()
	writer := bufio.NewWriter(newFile)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	err = writer.Flush()
	if err != nil {
		log.Fatal(err)
	}
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
				addToPath(c.Args().First(), true)
			},
		},
		{
			Name:      "append",
			ShortName: "a",
			Usage:     "Appends the directory to your PATH variable",
			Action: func(c *cli.Context) {
				addToPath(c.Args().First(), false)
			},
		},
		{
			Name:      "which",
			ShortName: "w",
			Usage:     "Shows the the different path entry where the program appears. Like which, but not just with the first location",
			Action: func(c *cli.Context) {
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

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
	//"github.com/codegangsta/cli"
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
		j := 0
		for scanner.Scan() {
			if search.Match([]byte(scanner.Text())) {
				location.Path = v
				location.Line = j + 1
				goto done
			}
			j++
		}
	}

	log.Fatal("couldn't find $PATH")

done:
	return location
}

func main() {
	/*app := cli.NewApp()
	  	app.Name = "go_path_manager"
	  	app.Usage = "manage your path variable"
	  	app.Commands = []cli.Command{
	  		{
	  			Name:      "list",
	    		ShortName: "l",
	    		Usage:     "list your path",
	    		Action: func(c *cli.Context) {
	      			println("added task: ", c.Args().First())
	    		}
	  		},
	  		{
	  			Name:      "prepend",
	    		ShortName: "p",
	    		Usage:     "prepend a directory to your path",
	    		Action: func(c *cli.Context) {
	      			println("added task: ", c.Args().First())
	    		}
	  		},
	  		{
	  			Name:      "append",
	    		ShortName: "a",
	    		Usage:     "append a directory to your path",
	    		Action: func(c *cli.Context) {
	      			println("added task: ", c.Args().First())
	    		}
	  		},
	  		{
	  			Name:      "which",
	    		ShortName: "w",
	    		Usage:     "show the location of a program in your path",
	    		Action: func(c *cli.Context) {
	      			println("added task: ", c.Args().First())
	    		}
	  		}
	  	}
	  	app.Run(os.Args)*/
	fmt.Println(getPathLocation())
}

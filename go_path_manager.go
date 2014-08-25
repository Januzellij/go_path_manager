package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"
	//"github.com/codegangsta/cli"
)

type Config struct {
	Zsh, Bash, Sh, Ksh []string
}

type Location struct {
	Path string
	Line int
}

func getPathLocation() Location {
	file, err := os.Open("./config.json")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	data := make([]byte, 600)
	n, err := file.Read(data)
	if err != nil {
		log.Fatalln(err)
	}

	var config Config
	err = json.Unmarshal(data[:n], &config)
	if err != nil {
		log.Fatalln(err)
	}

	configDict := map[string][]string{
		"zsh":  config.Zsh,
		"bash": config.Bash,
		"sh":   config.Sh,
		"ksh":  config.Ksh}

	usercmd := exec.Command("whoami")
	rawUser, err := usercmd.Output()
	if err != nil {
		log.Fatalln(err)
	}

	username := strings.Replace(string(rawUser), "\n", "", -1)
	infocmd := exec.Command("finger", username)
	info, err := infocmd.Output()
	if err != nil {
		log.Fatalln(err)
	}

	shellRe := regexp.MustCompile(`Shell: \/bin\/([a-zA-Z\/].*)`)
	shellStr := shellRe.FindString(string(info))
	if shellStr == "" {
		log.Fatalln("Couldn't find user shell")
	}

	shell := strings.Split(shellStr, "/")[2]
	paths := configDict[shell]

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

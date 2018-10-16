package main

import (
	"os"
	"fmt"
	"log"
	"bufio"
	"regexp"
	"strings"
	"os/exec"

)

type Info struct {
	props map[string]string
	stages []string
	externalRepo bool
}

func main() {
	info := findInfo(os.Args[1:])

	switch info.props["action"] {
	case "build":
		dockerPull(info)
		dockerBuild(info)
	case "push":
		dockerPush(info)
	}
}

func findInfo(args []string) Info {
	info := Info{
		props: map[string]string{
			"cachePrefix": "cache-",
		},
	}

	rtag, _ := regexp.Compile(" (-t|--tag)[ =]([^ ]+)")

	// Action
	info.props["action"] = args[1]

	// Tags
	switch info.props["action"] {
	case "build":
		info.props["targetTag"] = rtag.FindStringSubmatch(strings.Join(args, " "))[2]
	case "push":
		info.props["targetTag"] = args[len(args)-1]
	}

	r, _ := regexp.Compile("([^:]*:)?(.*)")
	info.props["tagTemplate"] = r.ReplaceAllString(info.props["targetTag"], "$1%s")

	info.externalRepo = strings.Contains(info.props["targetTag"], "/")

	if info.props["action"] == "build" {
		cmd := strings.Join(append(
			args[0:len(args)-1], // base command
			"%s", // template
			args[len(args)-1], // build directory
		), " ")
		info.props["buildTemplate"] = rtag.ReplaceAllString(cmd, "")
	}

	// Dockerfile
	switch info.props["action"] {
	case "build":
		info.props["dockerfile"] = findDockerfile(args[len(args)-1])
	case "push":
		info.props["dockerfile"] = findDockerfile(".")
	}

	// Stages
	info.stages = getStages(info.props["dockerfile"])

	return info
}

func dockerPull(info Info) {
	if info.externalRepo {
		for _, stage := range info.stages {
			image := fmt.Sprintf(info.props["tagTemplate"], info.props["cachePrefix"]+stage)
			execCmd(fmt.Sprintf("docker pull %s", image))
		}
	}
	execCmd(fmt.Sprintf("docker pull %s", info.props["targetTag"]))
}

func dockerBuild(info Info) {
	cacheFrom := ""
	for _, stage := range info.stages {
		cacheTag := fmt.Sprintf(info.props["tagTemplate"], info.props["cachePrefix"]+stage)
		cacheFrom += fmt.Sprintf(" --cache-from %s", cacheTag)

		// Set new cache tag
		extraArgs := fmt.Sprintf("-t %s", cacheTag)
		// Set previous cache from
		extraArgs += cacheFrom
		// Target to build
		extraArgs += fmt.Sprintf(" --target %s", stage)
		execCmd(fmt.Sprintf(info.props["buildTemplate"], extraArgs))
	}

	extraMainArgs := fmt.Sprintf("-t %s", info.props["targetTag"])
	cacheFrom += fmt.Sprintf(" --cache-from %s", info.props["targetTag"])
	extraMainArgs += cacheFrom
	execCmd(fmt.Sprintf(info.props["buildTemplate"], extraMainArgs))
}

func dockerPush(info Info) {
	execCmd(fmt.Sprintf("docker push %s", info.props["targetTag"]))
	for _, stage := range info.stages {
		execCmd(fmt.Sprintf("docker push %s", fmt.Sprintf(info.props["tagTemplate"], info.props["cachePrefix"]+stage)))
	}
}

func execCmd(cmdString string) {
	fmt.Printf("\n# %s\n", cmdString)
	cmdSlice := strings.Split(cmdString, " ")
	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func findDockerfile(path string) string {
	fi, err := os.Stat(path)
	if (err != nil) {
		log.Fatal(err)
	}
	if fi.Mode().IsRegular() {
		return path
	}
	if fi.Mode().IsDir() {
		return findDockerfile(path+"/Dockerfile")
	}

	log.Fatal("Could not find Dockerfile")
	return ""
}

func getStages(dockerfile string) []string {
	file, err := os.Open(dockerfile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	r, _ := regexp.Compile("(?i)^FROM\\s+.*\\s+AS\\s+(.*)")

	scanner := bufio.NewScanner(file)
	stages := []string{}
	for scanner.Scan() {
		matches := r.FindStringSubmatch(scanner.Text())
		if (len(matches) == 2) {
			stages = append(stages, matches[1])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return stages
}
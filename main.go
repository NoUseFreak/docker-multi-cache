package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

type Info struct {
	props        map[string]string
	stages       []string
	externalRepo bool
}

func main() {
	// Prepend docker if the command got replaced
	args := os.Args[1:]
	if len(args) == 0 || args[0] != "docker" {
		args = append([]string{"docker"}, args...)
	}

	if len(args) == 1 {
		os.Exit(execCmd(strings.Join(args, " ")))
		return
	}

	switch args[1] {
	case "build":
		info := findInfo(args)
		dockerPull(info)
		os.Exit(dockerBuild(info))
	case "push":
		info := findInfo(args)
		os.Exit(dockerPush(info))
	default:
		os.Exit(execCmd(strings.Join(args, " ")))
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
			"%s",              // template
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

func dockerPull(info Info) int {
	for _, stage := range info.stages {
		image := fmt.Sprintf(info.props["tagTemplate"], info.props["cachePrefix"]+stage)
		execCmd(fmt.Sprintf("docker pull %s", image))
	}
	return execCmd(fmt.Sprintf("docker pull %s", info.props["targetTag"]))
}

func dockerBuild(info Info) int {
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
	return execCmd(fmt.Sprintf(info.props["buildTemplate"], extraMainArgs))
}

func dockerPush(info Info) int {
	exitCode := execCmd(fmt.Sprintf("docker push %s", info.props["targetTag"]))
	for _, stage := range info.stages {
		execCmd(fmt.Sprintf("docker push %s", fmt.Sprintf(info.props["tagTemplate"], info.props["cachePrefix"]+stage)))
	}
	return exitCode
}

func execCmd(cmdString string) int {
	fmt.Printf("\n# %s\n", cmdString)
	if os.Getenv("DEBUG") == "" {
		cmdSlice := strings.Split(cmdString, " ")
		cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				ws := exitError.Sys().(syscall.WaitStatus)
				return ws.ExitStatus()
			}
			return 1
		}
	}

	return 0
}

func findDockerfile(path string) string {
	fi, err := os.Stat(path)
	if err != nil {
		fmt.Println("Could not find Dockerfile")
		return ""
	}
	if fi.Mode().IsRegular() {
		return path
	}
	if fi.Mode().IsDir() {
		return findDockerfile(path + "/Dockerfile")
	}

	fmt.Println("Could not find Dockerfile")
	return ""
}

func getStages(dockerfile string) []string {
	file, err := os.Open(dockerfile)
	if err != nil {
		return []string{}
	}
	defer file.Close()

	r, _ := regexp.Compile("(?i)^FROM\\s+.*\\s+AS\\s+(.*)")

	scanner := bufio.NewScanner(file)
	stages := []string{}
	for scanner.Scan() {
		matches := r.FindStringSubmatch(scanner.Text())
		if len(matches) == 2 {
			stages = append(stages, matches[1])
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return []string{}
	}

	return stages
}

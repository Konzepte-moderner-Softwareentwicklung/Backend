package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"time"

	v "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/version"
)

func GetInfo() v.Info {
	out := gitOutput("describe", "--tags", "--always")
	version := strings.TrimPrefix(out, "v")
	commit := gitOutput("rev-parse", "--short", "HEAD")
	branch := gitOutput("rev-parse", "--abbrev-ref", "HEAD")
	return v.Info{
		Version:   version,
		Commit:    commit,
		Branch:    branch,
		BuildTime: time.Now().String(),
	}
}

func gitOutput(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func main() {
	// just get some info about the build and print to stdout as json
	file, err := os.Create("version.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	json.NewEncoder(file).Encode(GetInfo())
}

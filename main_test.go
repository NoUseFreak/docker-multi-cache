package main

import (
	"reflect"
	"testing"
)

func TestFindDockerfileExact(t *testing.T) {
	existing := "tests/Dockerfile"
	path := findDockerfile(existing)

	if existing != path {
		t.Errorf("Expected %s but got %s", existing, path)
	}
}

func TestFindDockerfileDirectory(t *testing.T) {
	existing := "tests/Dockerfile"
	path := findDockerfile("tests")

	if existing != path {
		t.Errorf("Expected %s but got %s", existing, path)
	}
}

func TestGetStages(t *testing.T) {
	stages := getStages("tests/Dockerfile")

	expected := []string{"base", "base2"}
	if !reflect.DeepEqual(stages, expected) {
		t.Errorf("Expected %v but got %v as stages", expected, stages)
	}
}

func TestFindInfoPull(t *testing.T) {
	info := findInfo([]string{"docker", "pull"})

	if _, ok := info.props["action"]; !ok {
		t.Errorf("Did not find action")
	}
	if info.props["action"] != "pull" {
		t.Errorf("Expected 'pull' but got %s as action", info.props["action"])
	}
}
func TestFindInfoTag(t *testing.T) {
	info := findInfo([]string{"docker", "tag"})

	if _, ok := info.props["action"]; !ok {
		t.Errorf("Did not find action")
	}
	if info.props["action"] != "tag" {
		t.Errorf("Expected 'tag' but got %s as action", info.props["action"])
	}
}

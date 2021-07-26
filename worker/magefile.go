// +build mage

package main

import (
	"encoding/json"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/magefile/mage/target"
)

type modInfo struct {
	Path      string
	Version   string
	Time      time.Time
	Dir       string
	GoMod     string
	GoVersion string
}

// regenerate protobuf
func Proto() error {
	cmd := exec.Command("go", "list", "-m", "-json", "github.com/livekit/protocol")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	info := modInfo{}
	if err = json.Unmarshal(out, &info); err != nil {
		return err
	}
	protoDir := info.Dir
	updated, err := target.Path(
		"proto/livekit_recording_internal.pb.go",
		protoDir+"/livekit_recording_internal.proto",
	)
	if err != nil {
		return err
	}
	if !updated {
		return nil
	}

	fmt.Println("generating protobuf")
	goOut := "proto"
	if err := os.MkdirAll(goOut, 0755); err != nil {
		return err
	}

	protoc, err := getToolPath("protoc")
	if err != nil {
		return err
	}
	protocGoPath, err := getToolPath("protoc-gen-go")
	if err != nil {
		return err
	}

	// generate internal
	cmd = exec.Command(protoc,
		"--go_out", goOut,
		"--go_opt=paths=source_relative",
		"--plugin=go="+protocGoPath,
		"-I="+protoDir,
		protoDir+"/livekit_recording_internal.proto",
	)
	connectStd(cmd)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// helpers

func getToolPath(name string) (string, error) {
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}
	// check under gopath
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	p := filepath.Join(gopath, "bin", name)
	if _, err := os.Stat(p); err != nil {
		return "", err
	}
	return p, nil
}

func connectStd(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

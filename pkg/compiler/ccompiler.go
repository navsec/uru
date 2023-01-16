// modified from https://github.com/BishopFox/sliver/blob/5bcfa4c249341e9c9032abcaaf1d4cf459e20059/server/gogo/go.go

package compiler

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
)

type CConfig struct {
	ProjectDir     string
	TargetOs       string
	TargetCompiler string
	buildDll       bool
	OutDir         string
	CompileFlags   []string
	ExportDefPath  string
	ArtifactList   []string

	Env []string
}

func (c *CConfig) GetExportNames(extension string) string {
	return ""
}

func (c *CConfig) GetDebugImports() []string {
	return []string{}
}

func NewEmptyCConfig() *CConfig {
	config := &CConfig{}
	return config
}

func (c *CConfig) PrepareBuild(buildData BuildData) error {

	var targetCompiler string = "/usr/local/bin/x86_64-w64-mingw32-gcc"
	var compileFlags []string
	if contains(buildData.Imports, "iostream") {
		// iostream is a c++ library
		targetCompiler = "/usr/local/bin/x86_64-w64-mingw32-g++"

		// this is a lazy was to solve several problems
		// conversion from const char * to unsigned char * (in c++ string literals are const char arrays and
		// we nee this flag to convert ot unsigned char array, there are other ways to do, but this is easy)
		compileFlags = append(compileFlags, "-fpermissive")
	}

	// -static needed because we cant assume mingw will be on the target system
	compileFlags = append(compileFlags, "--static")

	// -static-libgcc and -static-libstdc++ needed to link the C and C++ standard libraries statically and
	// remove the need to carry around any separate copies of those.
	compileFlags = append(compileFlags, "-static-libstdc++", "-static-libgcc")

	if buildData.BuildMode == "c-shared" {
		compileFlags = append(compileFlags, "-shared")
	}

	c.ProjectDir = buildData.DirPath
	c.TargetOs = buildData.TargetOs
	c.TargetCompiler = targetCompiler
	c.CompileFlags = compileFlags
	c.ArtifactList = buildData.ArtifactList

	return nil
}

func (c *CConfig) Build(payload, dest string) ([]byte, error) {

	// add the payload to the compiler arguments
	for _, flag := range c.CompileFlags {

		c.Env = append(c.Env, flag)
	}

	// remove abs path from dest var
	c.Env = append(c.Env, []string{fmt.Sprintf("-o %s", filepath.Base(dest))}...)

	// format arguments to be used by compiler
	c.Env = append(c.Env, []string{fmt.Sprintf("%s", payload)}...)

	if contains(c.ArtifactList, "dllforward") {
		c.Env = append(c.Env, []string{fmt.Sprintf("%s", "../../data/templates/c/evasions/dllforward/example.def")}...)
	}

	// create command (dont run it yet)
	cmd := exec.Command(c.TargetCompiler, c.Env...)

	cmd.Dir = c.ProjectDir

	cmd.Env = c.Env

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// run command, gather output
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error %s: %s", err, string(stderr.Bytes()))
	}

	return stdout.Bytes(), err
}

func (c *CConfig) IsTypeSupported(t string) (string, string, error) {

	switch t {
	case "exe":
		return "exe", "", nil
	case "dll":
		return "dll", "c-shared", nil
	default:
		return "", "", fmt.Errorf("unsupported executable type")
	}
}

// https://play.golang.org/p/Qg_uv_inCek
// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

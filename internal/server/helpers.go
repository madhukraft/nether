package server

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var targetOS = runtime.GOOS
var targetArch = runtime.GOARCH

func SetTarget(os, arch string) bool {
	if os != "" {
		normalized := os
		if normalized == "macos" {
			normalized = "darwin"
		}
		switch normalized {
		case "darwin", "linux", "windows":
			targetOS = normalized
		default:
			return false
		}
	}
	if arch != "" {
		targetArch = arch
	}
	return true
}

func TargetOS() string {
	return targetOS
}

func bundledJavaDir() string {
	if targetOS == "darwin" {
		return filepath.Join("java", "Contents", "Home")
	}
	return "java"
}

func bundledJavaPath() string {
	if targetOS == "windows" {
		return filepath.Join("java", "bin", "java.exe")
	}
	return filepath.Join(bundledJavaDir(), "bin", "java")
}

func bundledJavaRef() string {
	if targetOS == "windows" {
		return filepath.Join("java", "bin", "java.exe")
	}
	return "./" + filepath.Join(bundledJavaDir(), "bin", "java")
}

func bundledJavaHome() string {
	return bundledJavaDir()
}

func bundledJavaExists() bool {
	_, err := os.Stat(bundledJavaPath())
	return err == nil
}

func IsServerDirectory(dir string) bool {
	entries := []string{
		"eula.txt",
		"server.properties",
		"run.sh",
		"run.bat",
	}
	for _, name := range entries {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*.jar"))
	if err == nil && len(matches) > 0 {
		return true
	}
	return false
}

func patchScriptJava(path string, mode os.FileMode) error {
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var javaRef string
	if strings.HasSuffix(path, ".bat") {
		javaRef = "java\\bin\\java.exe"
	} else {
		javaRef = bundledJavaRef()
	}

	content := string(data)

	hadCRLF := strings.Contains(content, "\r\n")
	if hadCRLF {
		content = strings.ReplaceAll(content, "\r\n", "\n")
	}

	content = strings.Replace(content, "exec java", "exec "+javaRef, 1)
	if strings.HasPrefix(content, "java ") {
		content = strings.Replace(content, "java ", javaRef+" ", 1)
	}
	content = strings.ReplaceAll(content, "\njava ", "\n"+javaRef+" ")
	content = strings.ReplaceAll(content, "\njava\t", "\n"+javaRef+"\t")

	if hadCRLF {
		content = strings.ReplaceAll(content, "\n", "\r\n")
	}

	return os.WriteFile(path, []byte(content), mode)
}

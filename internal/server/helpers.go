package server

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func bundledJavaPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join("java", "bin", "java.exe")
	}
	return filepath.Join("java", "bin", "java")
}

func bundledJavaRef() string {
	if runtime.GOOS == "windows" {
		return filepath.Join("java", "bin", "java.exe")
	}
	return "./" + filepath.Join("java", "bin", "java")
}

func bundledJavaExists() bool {
	_, err := os.Stat(bundledJavaPath())
	return err == nil
}

func patchScriptJava(path string, mode os.FileMode) error {
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	javaRef := bundledJavaRef()
	content := string(data)

	content = strings.Replace(content, "exec java", "exec "+javaRef, 1)
	if strings.HasPrefix(content, "java ") {
		content = strings.Replace(content, "java ", javaRef+" ", 1)
	}
	content = strings.ReplaceAll(content, "\njava ", "\n"+javaRef+" ")
	content = strings.ReplaceAll(content, "\njava\t", "\n"+javaRef+"\t")

	return os.WriteFile(path, []byte(content), mode)
}

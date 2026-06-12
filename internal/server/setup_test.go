package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteEula(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	if err := WriteEula(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("eula.txt")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "eula=true") {
		t.Errorf("expected eula=true, got: %s", content)
	}
	if !strings.Contains(content, "MinecraftEULA") {
		t.Errorf("expected EULA reference, got: %s", content)
	}
}

func TestWriteServerProperties(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	if err := WriteServerProperties(25565); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("server.properties")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "server-port=25565\n" {
		t.Errorf("unexpected content: %q", string(data))
	}
}

func TestWriteStartScriptLinux(t *testing.T) {
	origOS := targetOS
	targetOS = "linux"
	defer func() { targetOS = origOS }()

	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	if err := WriteStartScript("1G", "2G"); err != nil {
		t.Fatal(err)
	}

	shData, err := os.ReadFile("run.sh")
	if err != nil {
		t.Fatal(err)
	}
	shContent := string(shData)
	if !strings.Contains(shContent, "-Xms1G") {
		t.Errorf("run.sh missing -Xms1G: %s", shContent)
	}
	if !strings.Contains(shContent, "-Xmx2G") {
		t.Errorf("run.sh missing -Xmx2G: %s", shContent)
	}
	if !strings.HasPrefix(shContent, "#!/usr/bin/env sh\n") {
		t.Errorf("run.sh should start with shebang, got: %q", shContent[:40])
	}

	if _, err := os.Stat("run.bat"); !os.IsNotExist(err) {
		t.Errorf("run.bat should not exist on linux target")
	}
}

func TestWriteStartScriptWindows(t *testing.T) {
	origOS := targetOS
	targetOS = "windows"
	defer func() { targetOS = origOS }()

	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	if err := WriteStartScript("1G", "2G"); err != nil {
		t.Fatal(err)
	}

	batData, err := os.ReadFile("run.bat")
	if err != nil {
		t.Fatal(err)
	}
	batContent := string(batData)
	if !strings.Contains(batContent, "-Xms1G") {
		t.Errorf("run.bat missing -Xms1G: %s", batContent)
	}
	if !strings.Contains(batContent, "-Xmx2G") {
		t.Errorf("run.bat missing -Xmx2G: %s", batContent)
	}
	if !strings.HasPrefix(batContent, "@echo off\r\n") {
		t.Errorf("run.bat should start with @echo off CRLF, got: %q", batContent[:25])
	}

	if _, err := os.Stat("run.sh"); !os.IsNotExist(err) {
		t.Errorf("run.sh should not exist on windows target")
	}
}

func TestWriteUserJVMArgs(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	if err := WriteUserJVMArgs("1G", "4G"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("user_jvm_args.txt")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "-Xms1G") {
		t.Errorf("missing -Xms1G: %s", content)
	}
	if !strings.Contains(content, "-Xmx4G") {
		t.Errorf("missing -Xmx4G: %s", content)
	}
	if !strings.Contains(content, "-XX:+UseG1GC") {
		t.Errorf("missing default JVM flag: %s", content)
	}
	if !strings.HasSuffix(strings.TrimSpace(content), "-Daikars.new.flags=true") {
		t.Errorf("expected default flags to end content, got: %s", content)
	}
}

func TestWriteConfig(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	if err := WriteConfig("fabric", "1.21.1"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("nether.toml")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `type = "fabric"`) {
		t.Errorf("missing type: %s", content)
	}
	if !strings.Contains(content, `version = "1.21.1"`) {
		t.Errorf("missing version: %s", content)
	}
}

func TestWriteStartScriptUsesBundledJavaRef(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	origOS := targetOS

	t.Run("linux", func(t *testing.T) {
		targetOS = "linux"
		defer func() { targetOS = origOS }()

		if err := WriteStartScript("1G", "2G"); err != nil {
			t.Fatal(err)
		}

		shData, err := os.ReadFile("run.sh")
		if err != nil {
			t.Fatal(err)
		}

		javaRef := filepath.ToSlash(bundledJavaRef())
		if !strings.Contains(string(shData), javaRef) {
			t.Errorf("run.sh should use bundledJavaRef %q, got: %s", javaRef, string(shData))
		}
	})

	t.Run("windows", func(t *testing.T) {
		targetOS = "windows"

		if err := WriteStartScript("1G", "2G"); err != nil {
			t.Fatal(err)
		}

		batData, err := os.ReadFile("run.bat")
		if err != nil {
			t.Fatal(err)
		}
		expectedBatJava := `java\bin\java.exe`
		if !strings.Contains(string(batData), expectedBatJava) {
			t.Errorf("run.bat should use %q, got: %s", expectedBatJava, string(batData))
		}
	})
}

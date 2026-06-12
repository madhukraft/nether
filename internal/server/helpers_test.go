package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetTarget(t *testing.T) {
	origOS := targetOS
	origArch := targetArch
	defer func() {
		targetOS = origOS
		targetArch = origArch
	}()

	tests := []struct {
		name    string
		os      string
		arch    string
		wantOK  bool
		wantOS  string
		wantErr bool
	}{
		{"linux", "linux", "", true, "linux", false},
		{"macos normalized to darwin", "macos", "", true, "darwin", false},
		{"darwin", "darwin", "", true, "darwin", false},
		{"windows", "windows", "", true, "windows", false},
		{"invalid OS rejected", "freebsd", "", false, origOS, false},
		{"arch only, os empty", "", "arm64", true, origOS, false},
		{"both set", "linux", "arm64", true, "linux", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetOS = origOS
			targetArch = origArch
			got := SetTarget(tt.os, tt.arch)
			if got != tt.wantOK {
				t.Errorf("SetTarget(%q, %q) = %v; want %v", tt.os, tt.arch, got, tt.wantOK)
			}
			if tt.os != "" && tt.wantOK {
				if targetOS != tt.wantOS {
					t.Errorf("targetOS = %q; want %q", targetOS, tt.wantOS)
				}
			}
		})
	}
}

func TestBundledJavaDir(t *testing.T) {
	origOS := targetOS
	defer func() { targetOS = origOS }()

	targetOS = "darwin"
	if d := bundledJavaDir(); d != filepath.Join("java", "Contents", "Home") {
		t.Errorf("on darwin expected java/Contents/Home, got %q", d)
	}

	targetOS = "linux"
	if d := bundledJavaDir(); d != "java" {
		t.Errorf("on linux expected java, got %q", d)
	}

	targetOS = "windows"
	if d := bundledJavaDir(); d != "java" {
		t.Errorf("on windows expected java, got %q", d)
	}
}

func TestBundledJavaPath(t *testing.T) {
	origOS := targetOS
	defer func() { targetOS = origOS }()

	targetOS = "darwin"
	p := bundledJavaPath()
	if !strings.HasSuffix(p, "java") {
		t.Errorf("on darwin expected .../java, got %q", p)
	}
	if !strings.Contains(p, "Contents/Home") {
		t.Errorf("on darwin expected Contents/Home, got %q", p)
	}

	targetOS = "linux"
	p = bundledJavaPath()
	if !strings.HasSuffix(p, "java") || strings.HasPrefix(p, "./") {
		t.Errorf("on linux expected .../java without ./ prefix, got %q", p)
	}

	targetOS = "windows"
	p = bundledJavaPath()
	if !strings.HasSuffix(p, "java.exe") {
		t.Errorf("on windows expected .../java.exe, got %q", p)
	}
}

func TestBundledJavaRef(t *testing.T) {
	origOS := targetOS
	defer func() { targetOS = origOS }()

	targetOS = "darwin"
	ref := bundledJavaRef()
	if !strings.HasPrefix(ref, "./") {
		t.Errorf("on darwin expected ./ prefix, got %q", ref)
	}
	if !strings.Contains(ref, "Contents/Home") {
		t.Errorf("on darwin expected Contents/Home, got %q", ref)
	}

	targetOS = "linux"
	ref = bundledJavaRef()
	if !strings.HasPrefix(ref, "./") {
		t.Errorf("on linux expected ./ prefix, got %q", ref)
	}
	if strings.Contains(ref, "Contents/Home") {
		t.Errorf("on linux should NOT have Contents/Home, got %q", ref)
	}

	targetOS = "windows"
	ref = bundledJavaRef()
	if strings.HasPrefix(ref, "./") {
		t.Errorf("on windows should NOT have ./ prefix, got %q", ref)
	}
	if !strings.HasSuffix(ref, "java.exe") {
		t.Errorf("on windows expected java.exe, got %q", ref)
	}
}

func TestPatchScriptJava(t *testing.T) {
	origOS := targetOS
	defer func() { targetOS = origOS }()
	targetOS = "linux"

	t.Run("nonexistent file returns nil", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "run.sh")
		if err := patchScriptJava(path, 0755); err != nil {
			t.Fatalf("expected nil for nonexistent file, got: %v", err)
		}
	})

	t.Run("patches java command in shell script", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "run.sh")
		content := "#!/bin/bash\njava -Xmx2G -jar server.jar --nogui\n"
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatal(err)
		}

		if err := patchScriptJava(path, 0755); err != nil {
			t.Fatal(err)
		}

		data, _ := os.ReadFile(path)
		expectedJava := bundledJavaRef()
		if !strings.Contains(string(data), expectedJava) {
			t.Errorf("expected patched java ref %q, got: %s", expectedJava, string(data))
		}
		if strings.Contains(string(data), "\njava ") {
			t.Errorf("should not contain bare '\\njava ': %s", string(data))
		}
	})

	t.Run("preserves CRLF in bat file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "run.bat")
		content := "@echo off\r\njava -Xmx2G -jar server.jar --nogui\r\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		if err := patchScriptJava(path, 0644); err != nil {
			t.Fatal(err)
		}

		data, _ := os.ReadFile(path)
		if !strings.Contains(string(data), "\r\n") {
			t.Errorf("expected CRLF preserved, got: %q", string(data))
		}
		if !strings.Contains(string(data), `java\bin\java.exe`) {
			t.Errorf("expected bat java ref, got: %s", string(data))
		}
	})

	t.Run("patches exec java", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "run.sh")
		content := "#!/bin/bash\nexec java -Xmx2G -jar server.jar --nogui\n"
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatal(err)
		}

		if err := patchScriptJava(path, 0755); err != nil {
			t.Fatal(err)
		}

		data, _ := os.ReadFile(path)
		if !strings.Contains(string(data), "exec "+bundledJavaRef()) {
			t.Errorf("expected exec with patched java ref, got: %s", string(data))
		}
	})
}

func TestTargetOS(t *testing.T) {
	orig := targetOS
	defer func() { targetOS = orig }()

	targetOS = "linux"
	if TargetOS() != "linux" {
		t.Errorf("expected linux, got %q", TargetOS())
	}
}

func TestBundledJavaHomeDelegatesToDir(t *testing.T) {
	orig := targetOS
	defer func() { targetOS = orig }()

	targetOS = "darwin"
	if home := bundledJavaHome(); home != bundledJavaDir() {
		t.Errorf("bundledJavaHome() = %q, bundledJavaDir() = %q; expected equal", home, bundledJavaDir())
	}
}

func TestBundledJavaExists(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	if bundledJavaExists() {
		t.Error("expected false when java doesn't exist")
	}

	os.MkdirAll(filepath.Dir(bundledJavaPath()), 0755)
	os.WriteFile(bundledJavaPath(), []byte("fake"), 0755)

	if !bundledJavaExists() {
		t.Error("expected true when java exists")
	}
}

func TestIsServerDirectoryEmptyDir(t *testing.T) {
	dir := t.TempDir()
	if IsServerDirectory(dir) {
		t.Error("expected false for empty directory")
	}
}

func TestIsServerDirectoryEula(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "eula.txt"), []byte("eula=true"), 0644)
	if !IsServerDirectory(dir) {
		t.Error("expected true when eula.txt exists")
	}
}

func TestIsServerDirectoryServerProperties(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "server.properties"), []byte("port=25565"), 0644)
	if !IsServerDirectory(dir) {
		t.Error("expected true when server.properties exists")
	}
}

func TestIsServerDirectoryRunSh(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "run.sh"), []byte("#!/bin/bash"), 0644)
	if !IsServerDirectory(dir) {
		t.Error("expected true when run.sh exists")
	}
}

func TestIsServerDirectoryRunBat(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "run.bat"), []byte("@echo off"), 0644)
	if !IsServerDirectory(dir) {
		t.Error("expected true when run.bat exists")
	}
}

func TestIsServerDirectoryJar(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "server.jar"), []byte("fake jar"), 0644)
	if !IsServerDirectory(dir) {
		t.Error("expected true when .jar file exists")
	}
}

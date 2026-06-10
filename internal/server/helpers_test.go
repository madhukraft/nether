package server

import (
	"os"
	"path/filepath"
	"testing"
)

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

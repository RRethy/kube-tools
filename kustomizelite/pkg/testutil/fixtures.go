package testutil

import (
	"path/filepath"
	"runtime"
)

func FixturePath(paths ...string) string {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(filepath.Dir(filename)))

	allPaths := append([]string{root, "fixtures"}, paths...)
	return filepath.Join(allPaths...)
}

func ValidFixturePath(name string) string {
	return FixturePath("valid", name, "kustomization.yaml")
}

func InvalidFixturePath(name string) string {
	return FixturePath("invalid", name, "kustomization.yaml")
}

func MultiDocFixturePath(name string) string {
	return FixturePath("multi-doc", name, "kustomization.yaml")
}

func ResourcePath(validDir, filename string) string {
	return FixturePath("valid", validDir, filename)
}

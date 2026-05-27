package stdlib_fs_test

import (
	"os"
	"strings"
	"testing"

	"github.com/bonkzero404/uddin-lang/interpreter"
	_ "github.com/bonkzero404/uddin-lang/stdlib/fs"
)

func runScript(t *testing.T, src string) string {
	t.Helper()
	var buf strings.Builder
	config := interpreter.TestConfig()
	config.Stdout = &buf
	prog, err := interpreter.ParseProgram([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	_, err = interpreter.Execute(prog, config)
	if err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestFSWriteAndRead(t *testing.T) {
	tmpFile := os.TempDir() + "/uddin_fs_test_write_read.txt"
	defer os.Remove(tmpFile)

	out := runScript(t, `
import "fs"
ok = fs.write("`+tmpFile+`", "hello stdlib fs")
print(ok)
content = fs.read("`+tmpFile+`")
print(content)
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected write to return true, got: %s", out)
	}
	if !strings.Contains(out, "hello stdlib fs") {
		t.Fatalf("expected file content, got: %s", out)
	}
}

func TestFSExists(t *testing.T) {
	tmpFile := os.TempDir() + "/uddin_fs_test_exists.txt"
	os.WriteFile(tmpFile, []byte("x"), 0644)
	defer os.Remove(tmpFile)

	out := runScript(t, `
import "fs"
print(fs.exists("`+tmpFile+`"))
print(fs.exists("/nonexistent/path/xyz"))
`)
	if !strings.Contains(out, "true") {
		t.Fatalf("expected true for existing file, got: %s", out)
	}
	if !strings.Contains(out, "false") {
		t.Fatalf("expected false for nonexistent path, got: %s", out)
	}
}

func TestFSPathJoin(t *testing.T) {
	out := runScript(t, `
import "fs"
result = fs.path_join("/tmp", "subdir", "file.txt")
print(result)
`)
	if !strings.Contains(out, "tmp") || !strings.Contains(out, "subdir") || !strings.Contains(out, "file.txt") {
		t.Fatalf("expected joined path, got: %s", out)
	}
}

func TestFSPathBasename(t *testing.T) {
	out := runScript(t, `
import "fs"
result = fs.path_basename("/path/to/file.txt")
print(result)
`)
	if !strings.Contains(out, "file.txt") {
		t.Fatalf("expected file.txt, got: %s", out)
	}
}

func TestFSPathExt(t *testing.T) {
	out := runScript(t, `
import "fs"
result = fs.path_ext("document.pdf")
print(result)
`)
	if !strings.Contains(out, ".pdf") {
		t.Fatalf("expected .pdf, got: %s", out)
	}
}

func TestFSGetcwd(t *testing.T) {
	out := runScript(t, `
import "fs"
cwd = fs.getcwd()
print(cwd)
`)
	out = strings.TrimSpace(out)
	if len(out) == 0 {
		t.Fatalf("expected non-empty cwd, got empty string")
	}
}

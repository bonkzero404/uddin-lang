package stdlib_fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bonkzero404/uddin-lang/interpreter"
)

type fsModule struct{}

func (m *fsModule) Name() string { return "fs" }

func (m *fsModule) Functions() map[string]interpreter.ModuleFunc {
	return map[string]interpreter.ModuleFunc{
		"read":          readFunc,
		"write":         writeFunc,
		"exists":        existsFunc,
		"size":          sizeFunc,
		"modified":      modifiedFunc,
		"permissions":   permissionsFunc,
		"mkdir":         mkdirFunc,
		"rmdir":         rmdirFunc,
		"list_dir":      listDirFunc,
		"copy":          copyFunc,
		"move":          moveFunc,
		"delete":        deleteFunc,
		"path_join":     pathJoinFunc,
		"path_dirname":  pathDirnameFunc,
		"path_basename": pathBasenameFunc,
		"path_ext":      pathExtFunc,
		"getcwd":        getcwdFunc,
		"chdir":         chdirFunc,
	}
}

func init() {
	interpreter.RegisterModule(&fsModule{})
}

// readFunc reads the content of a file.
func readFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.read(): expected 1 argument, got %d", len(args)))
	}
	filename, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.read(): filename must be a string"))
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return interpreter.Value(fmt.Errorf("fs.read(): %s", err.Error()))
	}
	return interpreter.Value(string(content))
}

// writeFunc writes content to a file.
func writeFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		return interpreter.Value(fmt.Errorf("fs.write(): expected 2 arguments, got %d", len(args)))
	}
	filename, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.write(): filename must be a string"))
	}
	content, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.write(): content must be a string"))
	}
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return interpreter.Value(fmt.Errorf("fs.write(): %s", err.Error()))
	}
	return interpreter.Value(true)
}

// existsFunc checks if a file or directory exists.
func existsFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.exists(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.exists(): path must be a string"))
	}
	_, err := os.Stat(path)
	return interpreter.Value(!os.IsNotExist(err))
}

// sizeFunc returns the size of a file in bytes.
func sizeFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.size(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.size(): path must be a string"))
	}
	info, err := os.Stat(path)
	if err != nil {
		return interpreter.Value(fmt.Errorf("fs.size(): %s", err.Error()))
	}
	return interpreter.Value(info.Size())
}

// modifiedFunc returns the last modification time of a file (Unix timestamp).
func modifiedFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.modified(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.modified(): path must be a string"))
	}
	info, err := os.Stat(path)
	if err != nil {
		return interpreter.Value(fmt.Errorf("fs.modified(): %s", err.Error()))
	}
	return interpreter.Value(info.ModTime().Unix())
}

// permissionsFunc returns the file permissions as a string.
func permissionsFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.permissions(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.permissions(): path must be a string"))
	}
	info, err := os.Stat(path)
	if err != nil {
		return interpreter.Value(fmt.Errorf("fs.permissions(): %s", err.Error()))
	}
	return interpreter.Value(info.Mode().String())
}

// mkdirFunc creates a directory. Optional second arg is recursive bool.
func mkdirFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 1 || len(args) > 2 {
		return interpreter.Value(fmt.Errorf("fs.mkdir(): expected 1 or 2 arguments"))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.mkdir(): path must be a string"))
	}
	recursive := false
	if len(args) == 2 {
		if r, ok := args[1].(bool); ok {
			recursive = r
		} else {
			return interpreter.Value(fmt.Errorf("fs.mkdir(): recursive flag must be a boolean"))
		}
	}
	var err error
	if recursive {
		err = os.MkdirAll(path, 0755)
	} else {
		err = os.Mkdir(path, 0755)
	}
	if err != nil {
		return interpreter.Value(fmt.Errorf("fs.mkdir(): %s", err.Error()))
	}
	return interpreter.Value(true)
}

// rmdirFunc removes a directory.
func rmdirFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.rmdir(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.rmdir(): path must be a string"))
	}
	if err := os.Remove(path); err != nil {
		return interpreter.Value(fmt.Errorf("fs.rmdir(): %s", err.Error()))
	}
	return interpreter.Value(true)
}

// listDirFunc lists the contents of a directory.
func listDirFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.list_dir(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.list_dir(): path must be a string"))
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return interpreter.Value(fmt.Errorf("fs.list_dir(): %s", err.Error()))
	}
	result := make([]interpreter.Value, len(entries))
	for i, entry := range entries {
		result[i] = interpreter.Value(entry.Name())
	}
	return interpreter.Value(result)
}

// copyFunc copies a file from source to destination.
func copyFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		return interpreter.Value(fmt.Errorf("fs.copy(): expected 2 arguments, got %d", len(args)))
	}
	src, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.copy(): source must be a string"))
	}
	dst, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.copy(): destination must be a string"))
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return interpreter.Value(fmt.Errorf("fs.copy(): %s", err.Error()))
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return interpreter.Value(fmt.Errorf("fs.copy(): %s", err.Error()))
	}
	defer dstFile.Close()
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return interpreter.Value(fmt.Errorf("fs.copy(): %s", err.Error()))
	}
	return interpreter.Value(true)
}

// moveFunc moves/renames a file.
func moveFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		return interpreter.Value(fmt.Errorf("fs.move(): expected 2 arguments, got %d", len(args)))
	}
	oldPath, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.move(): old path must be a string"))
	}
	newPath, ok := args[1].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.move(): new path must be a string"))
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return interpreter.Value(fmt.Errorf("fs.move(): %s", err.Error()))
	}
	return interpreter.Value(true)
}

// deleteFunc deletes a file.
func deleteFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.delete(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.delete(): path must be a string"))
	}
	if err := os.Remove(path); err != nil {
		return interpreter.Value(fmt.Errorf("fs.delete(): %s", err.Error()))
	}
	return interpreter.Value(true)
}

// pathJoinFunc joins path elements.
func pathJoinFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) == 0 {
		return interpreter.Value(fmt.Errorf("fs.path_join(): expected at least 1 argument"))
	}
	paths := make([]string, len(args))
	for i, arg := range args {
		if path, ok := arg.(string); ok {
			paths[i] = path
		} else {
			return interpreter.Value(fmt.Errorf("fs.path_join(): all arguments must be strings"))
		}
	}
	return interpreter.Value(filepath.Join(paths...))
}

// pathDirnameFunc returns the directory name of a path.
func pathDirnameFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.path_dirname(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.path_dirname(): path must be a string"))
	}
	return interpreter.Value(filepath.Dir(path))
}

// pathBasenameFunc returns the base name of a path.
func pathBasenameFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.path_basename(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.path_basename(): path must be a string"))
	}
	return interpreter.Value(filepath.Base(path))
}

// pathExtFunc returns the file extension.
func pathExtFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.path_ext(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.path_ext(): path must be a string"))
	}
	return interpreter.Value(filepath.Ext(path))
}

// getcwdFunc returns the current working directory.
func getcwdFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	cwd, err := os.Getwd()
	if err != nil {
		return interpreter.Value(fmt.Errorf("fs.getcwd(): %s", err.Error()))
	}
	return interpreter.Value(cwd)
}

// chdirFunc changes the current working directory.
func chdirFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(fmt.Errorf("fs.chdir(): expected 1 argument, got %d", len(args)))
	}
	path, ok := args[0].(string)
	if !ok {
		return interpreter.Value(fmt.Errorf("fs.chdir(): path must be a string"))
	}
	if err := os.Chdir(path); err != nil {
		return interpreter.Value(fmt.Errorf("fs.chdir(): %s", err.Error()))
	}
	return interpreter.Value(true)
}

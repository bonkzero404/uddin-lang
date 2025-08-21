package interpreter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// readFileFunc reads the content of a file
func readFileFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "read_file", args, 1)

	filename, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("read_file: filename must be a string"))
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return Value(fmt.Errorf("read_file: %s", err.Error()))
	}

	return Value(string(content))
}

// writeFileFunc writes content to a file
func writeFileFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "write_file", args, 2)

	filename, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("write_file: filename must be a string"))
	}

	content, ok := args[1].(string)
	if !ok {
		return Value(fmt.Errorf("write_file: content must be a string"))
	}

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return Value(fmt.Errorf("write_file: %s", err.Error()))
	}

	return Value(true)
}

// fileExistsFunc checks if a file or directory exists
// Usage: file_exists("path/to/file")
func fileExistsFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "file_exists", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("file_exists: path must be a string"))
	}
	_, err := os.Stat(path)
	return Value(!os.IsNotExist(err))
}

// fileSizeFunc returns the size of a file in bytes
// Usage: file_size("path/to/file")
func fileSizeFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "file_size", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("file_size: path must be a string"))
	}
	info, err := os.Stat(path)
	if err != nil {
		return Value(fmt.Errorf("file_size: %s", err.Error()))
	}
	return Value(info.Size())
}

// fileModifiedFunc returns the last modification time of a file
// Usage: file_modified("path/to/file")
func fileModifiedFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "file_modified", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("file_modified: path must be a string"))
	}
	info, err := os.Stat(path)
	if err != nil {
		return Value(fmt.Errorf("file_modified: %s", err.Error()))
	}
	return Value(info.ModTime().Unix())
}

// filePermissionsFunc returns the file permissions as a string
// Usage: file_permissions("path/to/file")
func filePermissionsFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "file_permissions", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("file_permissions: path must be a string"))
	}
	info, err := os.Stat(path)
	if err != nil {
		return Value(fmt.Errorf("file_permissions: %s", err.Error()))
	}
	return Value(info.Mode().String())
}

// mkdirFunc creates a directory
// Usage: mkdir("path/to/dir") or mkdir("path/to/dir", true) for recursive
func mkdirFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		return Value(fmt.Errorf("mkdir: expected 1 or 2 arguments"))
	}
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("mkdir: path must be a string"))
	}

	recursive := false
	if len(args) == 2 {
		if r, ok := args[1].(bool); ok {
			recursive = r
		} else {
			return Value(fmt.Errorf("mkdir: recursive flag must be a boolean"))
		}
	}

	var err error
	if recursive {
		err = os.MkdirAll(path, 0755)
	} else {
		err = os.Mkdir(path, 0755)
	}

	if err != nil {
		return Value(fmt.Errorf("mkdir: %s", err.Error()))
	}
	return Value(true)
}

// rmdirFunc removes a directory
// Usage: rmdir("path/to/dir")
func rmdirFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "rmdir", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("rmdir: path must be a string"))
	}
	err := os.Remove(path)
	if err != nil {
		return Value(fmt.Errorf("rmdir: %s", err.Error()))
	}
	return Value(true)
}

// listDirFunc lists the contents of a directory
// Usage: list_dir("path/to/dir")
func listDirFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "list_dir", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("list_dir: path must be a string"))
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return Value(fmt.Errorf("list_dir: %s", err.Error()))
	}

	result := make([]Value, len(entries))
	for i, entry := range entries {
		result[i] = Value(entry.Name())
	}
	return Value(result)
}

// copyFileFunc copies a file from source to destination
// Usage: copy_file("source", "destination")
func copyFileFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "copy_file", args, 2)
	src, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("copy_file: source must be a string"))
	}
	dst, ok := args[1].(string)
	if !ok {
		return Value(fmt.Errorf("copy_file: destination must be a string"))
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return Value(fmt.Errorf("copy_file: %s", err.Error()))
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return Value(fmt.Errorf("copy_file: %s", err.Error()))
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return Value(fmt.Errorf("copy_file: %s", err.Error()))
	}
	return Value(true)
}

// moveFileFunc moves/renames a file
// Usage: move_file("old_path", "new_path")
func moveFileFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "move_file", args, 2)
	oldPath, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("move_file: old path must be a string"))
	}
	newPath, ok := args[1].(string)
	if !ok {
		return Value(fmt.Errorf("move_file: new path must be a string"))
	}

	err := os.Rename(oldPath, newPath)
	if err != nil {
		return Value(fmt.Errorf("move_file: %s", err.Error()))
	}
	return Value(true)
}

// deleteFileFunc deletes a file
// Usage: delete_file("path/to/file")
func deleteFileFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "delete_file", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("delete_file: path must be a string"))
	}
	err := os.Remove(path)
	if err != nil {
		return Value(fmt.Errorf("delete_file: %s", err.Error()))
	}
	return Value(true)
}

// pathJoinFunc joins path elements
// Usage: path_join("dir", "subdir", "file.txt")
func pathJoinFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) == 0 {
		return Value(fmt.Errorf("path_join: expected at least 1 argument"))
	}

	paths := make([]string, len(args))
	for i, arg := range args {
		if path, ok := arg.(string); ok {
			paths[i] = path
		} else {
			return Value(fmt.Errorf("path_join: all arguments must be strings"))
		}
	}

	return Value(filepath.Join(paths...))
}

// pathDirnameFunc returns the directory name of a path
// Usage: path_dirname("/path/to/file.txt")
func pathDirnameFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "path_dirname", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("path_dirname: path must be a string"))
	}
	return Value(filepath.Dir(path))
}

// pathBasenameFunc returns the base name of a path
// Usage: path_basename("/path/to/file.txt")
func pathBasenameFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "path_basename", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("path_basename: path must be a string"))
	}
	return Value(filepath.Base(path))
}

// pathExtFunc returns the file extension
// Usage: path_ext("file.txt")
func pathExtFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "path_ext", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("path_ext: path must be a string"))
	}
	return Value(filepath.Ext(path))
}

// getcwdFunc returns the current working directory
// Usage: getcwd()
func getcwdFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "getcwd", args, 0)
	cwd, err := os.Getwd()
	if err != nil {
		return Value(fmt.Errorf("getcwd: %s", err.Error()))
	}
	return Value(cwd)
}

// chdirFunc changes the current working directory
// Usage: chdir("/new/directory")
func chdirFunc(interp *interpreter, pos Position, args []Value) Value {
	ensureNumArgs(pos, "chdir", args, 1)
	path, ok := args[0].(string)
	if !ok {
		return Value(fmt.Errorf("chdir: path must be a string"))
	}
	err := os.Chdir(path)
	if err != nil {
		return Value(fmt.Errorf("chdir: %s", err.Error()))
	}
	return Value(true)
}

package main

import (
    "bufio"
    "fmt"
    "io/fs"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

func backupDirectory(oldDirPath string) string {
    newDir := fmt.Sprintf("%s.backup", oldDirPath)
    cmd := exec.Command("cp", "--recursive", oldDirPath, newDir)
    cmd.Run()
    return newDir
}

func walkDirectory(runOnThisDirectory string, excludeDirectories map[string]bool) error {
    sb := strings.Builder{}
    err := filepath.Walk(runOnThisDirectory, func(path string, info fs.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() && excludeDirectories[info.Name()] {
            return filepath.SkipDir
        }

        if !info.IsDir() {
            fmt.Fprintf(os.Stderr, "INFO: On path: %s\n", path)
            f, err := os.OpenFile(path, os.O_RDWR, 0755)
            if err != nil {
                fmt.Fprintf(os.Stderr, "ERROR: Failed to open file at path: %s: %v\n", path, err)
                return err
            }
            defer f.Close()

            s := bufio.NewScanner(f)
            for s.Scan() {
                line := s.Text()
                sb.WriteString(strings.TrimSpace(line))
            }

            if err := f.Truncate(0); err != nil {
                fmt.Fprintf(os.Stderr, "ERROR: Failed to truncate file at path: %s: %v\n", path, err)
                return err
            }
            if _, err := f.Seek(0, 0); err != nil {
                fmt.Fprintf(os.Stderr, "ERROR: Failed to seek file at path: %s: %v\n", path, err)
                return err
            }
            if _, err := fmt.Fprintf(f, "%s", sb.String()); err != nil {
                fmt.Fprintf(os.Stderr, "ERROR: Failed to write to file at path: %s: %v\n", path, err)
                return err
            }
        }

        sb.Reset()
        return nil
    })
    if err != nil {
        return err
    }
    return nil
}

func main() {
    runOnThisDirectory := "web"
    excludeDirectories := map[string]bool{
        "imgs": true,
    }

    backedUpDirectory := backupDirectory(runOnThisDirectory)

    if err := walkDirectory(runOnThisDirectory, excludeDirectories); err != nil {
        fmt.Fprintf(os.Stderr, "ERROR: failed during filepath.Walk: %v", err)
        return
    }

    cmd := exec.Command("go", "build", "../dalennod/")
    if err := cmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "ERROR: failed to build using command: %s: %v", cmd.String(), err)
    }

    if err := os.RemoveAll(runOnThisDirectory); err != nil {
        fmt.Fprintf(os.Stderr, "ERROR: failed to remove directory: %s: %v", runOnThisDirectory, err)
    }
    if err := os.Rename(backedUpDirectory, runOnThisDirectory); err != nil {
        fmt.Fprintf(os.Stderr, "ERROR: renaming %s to %s: %v", backedUpDirectory, runOnThisDirectory, err)
    }
}

package gpm

import (
    "os"
    "fmt"
    "bytes"
    "strings"
    //"errors"
    "os/exec"
)


func ShellRun(cmd string) error {
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    execCmd := exec.Command("bash", "-c", cmd)
    execCmd.Stdout = &stdout
    execCmd.Stderr = &stderr
    err := execCmd.Run()
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", stderr.String())
        return err
    } else {
        fmt.Fprintf(os.Stdout, "%s\n", stdout.String())
        return nil
    }
}

func InstallDeb(pkgPath string) error {
    fmt.Println("Installing", pkgPath)

    insCmd := "sudo dpkg -i " + pkgPath
    err := ShellRun(insCmd)

    os.Remove(pkgPath)
    return err
}

func InstallOthPkg(cmds, pkgPath string) error {
    cmdsWithPkgPath := strings.Replace(cmds, "$PROG", pkgPath, -1)
    fmt.Println(cmdsWithPkgPath)
    err := ShellRun(cmdsWithPkgPath)

    return err
}

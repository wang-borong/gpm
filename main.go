package main

import (
    "os"
    "fmt"
    "log"
    "bufio"
    "strings"
    "os/exec"
    "path/filepath"
    "github.com/wbrn/gpm/lib"
)


func DownloadInstall(pkginfo string) {
    pkgCmd := strings.Split(pkginfo, "%")[0]
    pkgGhPath := strings.Split(pkginfo, "%")[1]
    pkgInsCmds := strings.Split(pkginfo, "%")[2]

    // if github owner and repo is none
    if pkgGhPath == "none" {
        fmt.Printf("runing %s ...", pkgInsCmds)
        gpm.ShellRun(pkgInsCmds)
    } else {
        ownerRepo := strings.Split(pkgGhPath, "/")

        oldVersion, err := exec.Command(pkgCmd, "--version").Output()
        if err != nil {
            fmt.Println(err)
            // just download it
            oldVersion = []byte("0.0.0")
        }

        pkgPath, err := gpm.DownloadLatestRelease(ownerRepo[0], ownerRepo[1], string(oldVersion[:]))
        if err != nil {
            fmt.Println(err)
            return
        }
        if pkgPath == "" {
            return
        }

        if pkgInsCmds == "deb" {
            err = gpm.InstallDeb(pkgPath)
            if err != nil {
                fmt.Println(err)
                return
            }
        } else {
            err = gpm.InstallOthPkg(pkgInsCmds, pkgPath)
            if err != nil {
                fmt.Println(err)
                return
            }
        }
    }
}

func main() {
    if len(os.Args) < 2 {
        home := os.Getenv("HOME")
        pkgListFile := filepath.Join(home, ".packages")
        file, err := os.Open(pkgListFile)
        if err != nil {
            log.Fatal(err)
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            if scanner.Text() == "" || scanner.Text()[0] == '#' {
                continue
            }
            DownloadInstall(scanner.Text())
        }

        if err := scanner.Err(); err != nil {
            log.Fatal(err)
        }
    } else {
        for _, pkginfo := range os.Args[1:] {
            DownloadInstall(pkginfo)
        }
    }
}

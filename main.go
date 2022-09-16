package main

import (
    "os"
    "fmt"
    "log"
    "bufio"
    "strings"
    "regexp"
    "os/exec"
    "path/filepath"
    "github.com/wbrn/gpm/lib"
)


func DownloadInstall(pkginfo string) {

    rePkgFmt1 := regexp.MustCompile(
        `^\s*([\w_/]+)\s*\(?@\)?\s*([\w-_/]+)\s*\(?!\)?\s*(.*)$`)

    rePkgFmt2 := regexp.MustCompile(
        `^\s*([\w_/]+)\s*\(?!\)?\s*(.*)$`)

    pkgInfoArr := rePkgFmt1.FindStringSubmatch(pkginfo)

    if pkgInfoArr == nil {
        pkgInfoArr = rePkgFmt2.FindStringSubmatch(pkginfo)
        if pkgInfoArr == nil {
            fmt.Println("No valid info for", pkginfo)
            return
        }
    }

    pkgName := pkgInfoArr[1]
    pkgGhPath := ""
    pkgInsCmds := pkgInfoArr[2]
    if len(pkgInfoArr) == 4 {
        pkgGhPath = pkgInfoArr[2]
        pkgInsCmds = pkgInfoArr[3]
    }

    if pkgGhPath == "" {
        fmt.Printf("runing %s ...", pkgInsCmds)
        gpm.ShellRun(pkgInsCmds)
    } else {
        ownerRepo := strings.Split(pkgGhPath, "/")

        oldVersion, err := exec.Command(pkgName, "--version").Output()
        if err != nil {
            fmt.Println(err)
            // just download it
            oldVersion = []byte("0.0.0")
        }

        pkgType := ""
        if strings.Contains(pkgInsCmds, "tar") {
            pkgType = "tar"
        } else if strings.Contains(pkgInsCmds, "deb") {
            pkgType = "deb"
        } else {
            pkgType = ""
        }
        pkgPath, err := gpm.DownloadLatestRelease(
            ownerRepo[0], ownerRepo[1], string(oldVersion[:]), pkgType)
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
        defer file.Close()
        if err != nil {
            log.Fatal(err)
        }

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            reBlank := regexp.MustCompile(`^\s*$`)
            reComment := regexp.MustCompile(`^\s*#+.*`)
            if reBlank.MatchString(scanner.Text()) ||
                reComment.MatchString(scanner.Text()) {
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

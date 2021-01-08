package main

import (
    "os"
    "fmt"
    "log"
    "bufio"
    "strings"
    "os/exec"
    "path/filepath"
    "wbr.com/gpm/lib"
)


func DownloadInstall(ghpath string) {
    pkgcmd := strings.Split(ghpath, ":")[0]
    owner_repo := strings.Split(strings.Split(ghpath, ":")[1], "/")

    oldVersion, err := exec.Command(pkgcmd, "--version").Output()
    if err != nil {
        fmt.Println(err)
        // just download it
        oldVersion = []byte("0.0.0")
    }

    pkg, err:= ghinstaller.DownloadLatestRelease(owner_repo[0], owner_repo[1], string(oldVersion[:]))
    if err != nil {
        fmt.Println(err)
        os.Exit(87)
    }
    err = ghinstaller.Install("dpkg", "-i", pkg)
    if err != nil {
        fmt.Println(err)
        os.Exit(87)
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
        for _, ghpath := range os.Args[1:] {
            DownloadInstall(ghpath)
        }
    }
}

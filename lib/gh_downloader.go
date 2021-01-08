package ghinstaller

import (
    "fmt"
    "io"
    "os"
    "regexp"
    "context"
    "strings"
    "strconv"
    "net/http"
    "path/filepath"
    "github.com/google/go-github/v33/github"
)

func CheckVersion(ghVersion, oldVersion string) bool {
    ghVerStr := strings.Replace(ghVersion, "v", "", -1)
    re := regexp.MustCompile(`\d{1,2}\.\d{1,2}\.\d{1,2}`)
    oldVerStr := re.FindString(oldVersion)

    if ghVerStr > oldVerStr {
        fmt.Printf("github version %s > local %s\n", ghVerStr, oldVerStr)
        return true
    } else {
        return false
    }
}

func DownloadLatestRelease(owner, repo, oldVersion string) (pkg string, err error) {
    var input string

    client := github.NewClient(nil)
    ctx := context.Background()
    release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
    if err != nil {
        return "", err
    }

    if !CheckVersion(*release.TagName, oldVersion) {
        fmt.Printf("%s %s is in latest version\n", repo, *release.TagName)
        return "", nil
    }

    for i, asset :=range release.Assets {
        fmt.Printf("[%02d] %s\n", i, *asset.Name)
    }
    fmt.Printf("Select your package: ")
    fmt.Scanln(&input)
    number, err := strconv.Atoi(input)

    appreAsset := release.Assets[number]
    pkgName := filepath.Join("/tmp", *appreAsset.Name)
    fmt.Println("Downloading", *appreAsset.Name)
    resp, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, *appreAsset.ID, http.DefaultClient)
    if err != nil {
        return "", err
    }

    if resp != nil {
        //Create a empty file
        file, err := os.Create(pkgName)
        if err != nil {
            return "", err
        }
        defer file.Close()

        //Write the bytes to the file
        _, err = io.Copy(file, resp)
        if err != nil {
            return "", err
        }
        defer resp.Close()
    }

    return pkgName, nil
}

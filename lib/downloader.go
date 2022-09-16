package gpm

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
    "github.com/elastic/go-sysinfo"
)

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(pkgPath string, url string) error {
    fmt.Printf("downloading %s ...", filepath.Base(pkgPath))
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(pkgPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func CheckVersion(ghVersion, oldVersion string) bool {
    //ghVerStr := strings.Replace(ghVersion, "v", "", -1)
    re := regexp.MustCompile(`\d{1,2}\.\d{1,2}(\.\d{1,2})?`)
    oldVerStr := re.FindString(oldVersion)
    ghVerStr := re.FindString(ghVersion)

    if ghVerStr > oldVerStr {
        fmt.Printf("github version %s > local %s\n", ghVerStr, oldVerStr)
        return true
    } else {
        return false
    }
}

func DownloadLatestTag(owner, repo, oldVersion string) error {
    client := github.NewClient(nil)
    ctx := context.Background()
    opt := &github.ListOptions{Page: 0}
    tags, _, err := client.Repositories.ListTags(ctx, owner, repo, opt)

    if err != nil {
        return err
    }
    latestTag := tags[0]
    if !CheckVersion(*latestTag.Name, oldVersion) {
        fmt.Printf("%s %s is in latest version\n", repo, *latestTag.Name)
        return nil
    }
    latestTagURL := *latestTag.TarballURL

    fileName := repo + "-" + filepath.Base(latestTagURL) + ".tar.gz"
    pkgPath := filepath.Join("/tmp", fileName)

    err = DownloadFile(pkgPath, latestTagURL)
    if err != nil {
        return err
    }

    return nil
}

func DownloadLatestRelease(owner, repo, oldVersion, pkgType string) (pkg string, err error) {
    var input string

    client := github.NewClient(nil)
    ctx := context.Background()
    release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
    if err != nil {
        if strings.Contains(err.Error(), "Not Found") {
            err = DownloadLatestTag(owner, repo, oldVersion)
            return "", err
        } else {
            return "", err
        }
    }

    ghVersion := *release.TagName
    if !CheckVersion(ghVersion, oldVersion) {
        fmt.Printf("%s %s is in latest version\n", repo, *release.TagName)
        return "", nil
    }

    host, _ := sysinfo.Host()
    // fmt.Println(host.Info().Architecture, host.Info().OS.Name)
    pkgFilter := make(map[string]int)

    for i, asset := range release.Assets {
        if strings.Contains(*asset.Name, pkgType) {
            if strings.Contains(host.Info().Architecture, "x86_64") {
                if strings.Contains(*asset.Name, "amd64") || 
                    strings.Contains(*asset.Name, "x86_64") {
                    if strings.Contains(host.Info().OS.Name, "GNU") {
                        if !strings.Contains(*asset.Name, "musl") {
                            pkgFilter[*asset.Name] = i
                        }
                    }
                }
            } else {
                if strings.Contains(*asset.Name, host.Info().Architecture) {
                    if strings.Contains(host.Info().OS.Name, "GNU") {
                        if !strings.Contains(*asset.Name, "musl") {
                            pkgFilter[*asset.Name] = i
                        }
                    }
                }
            }
        }
    }

    selPkg := 0
    for pn, i := range pkgFilter {
        fmt.Printf("[%02d] %s\n", i, pn)
        selPkg = i
    }

    if len(pkgFilter) > 1 {

        fmt.Printf("Select your package: ")
        fmt.Scanln(&input)
        number, err := strconv.Atoi(input)

        if err != nil {
            return "", err
        }
        selPkg = number
    }

    appreAsset := release.Assets[selPkg]
    pkgPath := filepath.Join("/tmp", *appreAsset.Name)

    if _, err := os.Stat(pkgPath); err == nil {
        return pkgPath, nil
    }

    fmt.Println("Downloading", *appreAsset.Name)
    resp, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, *appreAsset.ID, http.DefaultClient)
    if err != nil {
        return "", err
    }

    if resp != nil {
        //Create a empty file
        file, err := os.Create(pkgPath)
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

    return pkgPath, nil
}

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

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

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
    pkgName := filepath.Join("/tmp", fileName)

    err = DownloadFile(pkgName, latestTagURL)
    if err != nil {
        return err
    }

    return nil
}

func DownloadLatestRelease(owner, repo, oldVersion string) (pkg string, err error) {
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

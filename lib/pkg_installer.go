package ghinstaller

import (
    "os"
    "fmt"
    "strings"
    //"errors"
    "os/exec"
)

func Install(cmd string, arg string, pkg string) error {
    if !strings.HasSuffix(pkg, ".deb") {
        //return errors.New("Not a deb package")
        return nil
    }

    fmt.Println("Installing", pkg)
    // here we perform the pwd command.
    // we can store the output of this in our out variable
    // and catch any errors in err
    out, err := exec.Command("sudo", cmd, arg, pkg).Output()

    // as the out variable defined above is of type []byte we need to convert
    // this to a string or else we will see garbage printed out in our console
    // this is how we convert it to a string
    output := string(out[:])
    fmt.Println(output)

    // if there is an error with our execution
    // handle it here
    if err != nil {
        return err
    }

    os.Remove(pkg)
    return nil
}

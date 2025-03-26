package misc

import "fmt"

var versionString = "dev"
var commitHash = "dev"

func Version() {
	fmt.Printf("Version: %s\nCommit: %s\n", versionString, commitHash)
}

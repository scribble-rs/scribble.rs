// version holds the git that this version was built from. In development
// scearios, it will default to "dev".
package version

import (
	"regexp"
)

// Version of the application.
var (
	Version = "dev"
	Commit  = ""
)

func init() {
	// We expect to get a "dirt git tag" when deploying a test version that
	// we did not yet finalize. We'll take it apart to allow the frontend
	// to put the correct link to the repository.

	if Version != "dev" {
		// version-commit_count_after_version-hash
		hashRegex := regexp.MustCompile(`v.+?(?:-\d+?-g(.+?)(?:$|-))`)
		match := hashRegex.FindStringSubmatch(Version)
		if len(match) == 2 {
			Commit = match[1]
		}
	}
}

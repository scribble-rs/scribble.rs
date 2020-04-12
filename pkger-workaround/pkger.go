package pkger_workaround

import "path/filepath"

// Path offers a workaround for usage of resoruces packaged with pkger.
// See https://github.com/markbates/pkger/issues/86
func Path(path string) string {
    return filepath.Join("github.com/scribble-rs/scribble.rs:", path)
}

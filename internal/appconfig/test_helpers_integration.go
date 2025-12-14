//go:build integration
// +build integration

package appconfig

// OverrideConfigDir allows integration tests to point the loader at a temp config dir.
func (l *Loader) OverrideConfigDir(dir string) {
	l.SetConfigDir(dir)
}

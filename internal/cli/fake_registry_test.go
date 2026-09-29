package cli

// fake_registry_test.go provides the hermetic Registry double used by
// every test, so the suite never touches the real HKCU hive.

import "errors"

// FakeRegistry is an in-memory Registry for tests.
type FakeRegistry struct {
	registered    bool
	displayName   string
	ensureCalls   int
	ensureErr     error
	statusErr     error
	removeErr     error
	removedParent bool // guard flag: must stay false (parents never touched)
}

func (f *FakeRegistry) EnsureAUMID() error {
	f.ensureCalls++
	if f.ensureErr != nil {
		return f.ensureErr
	}
	f.registered = true
	if f.displayName == "" {
		f.displayName = AUMIDDisplayName
	}
	return nil
}

func (f *FakeRegistry) AUMIDStatus() (bool, string, error) {
	if f.statusErr != nil {
		return false, "", f.statusErr
	}
	if !f.registered || f.displayName == "" {
		return false, "AUMID registry key is missing", nil
	}
	return true, "registered (fake)", nil
}

func (f *FakeRegistry) RemoveAUMID() (bool, error) {
	if f.removeErr != nil {
		return false, f.removeErr
	}
	if !f.registered {
		return false, nil
	}
	f.registered = false
	f.displayName = ""
	return true, nil
}

var errTestBoom = errors.New("test registry failure")

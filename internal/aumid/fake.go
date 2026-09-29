package aumid

// fake.go provides the hermetic Registry double used by every test
// package, so the suite never touches the real HKCU hive.

// FakeRegistry is an in-memory Registry for tests.
type FakeRegistry struct {
	Registered    bool
	DisplayName   string
	EnsureCalls   int
	EnsureErr     error
	StatusErr     error
	RemoveErr     error
	RemovedParent bool // guard flag: must stay false (parents never touched)
}

func (f *FakeRegistry) EnsureAUMID() error {
	f.EnsureCalls++
	if f.EnsureErr != nil {
		return f.EnsureErr
	}
	f.Registered = true
	if f.DisplayName == "" {
		f.DisplayName = AUMIDDisplayName
	}
	return nil
}

func (f *FakeRegistry) AUMIDStatus() (bool, string, error) {
	if f.StatusErr != nil {
		return false, "", f.StatusErr
	}
	if !f.Registered || f.DisplayName == "" {
		return false, "AUMID registry key is missing", nil
	}
	return true, "registered (fake)", nil
}

func (f *FakeRegistry) RemoveAUMID() (bool, error) {
	if f.RemoveErr != nil {
		return false, f.RemoveErr
	}
	if !f.Registered {
		return false, nil
	}
	f.Registered = false
	f.DisplayName = ""
	return true, nil
}

package license

//go:generate mockery -all -inpkg

// License represents a software license.
type License struct {
	Name string // Name is a human-friendly name like "MIT License"
	SPDX string // SPDX ID of the license, blank if unknown or unavailable
	Text string // Text is the contents of the Licence
}

func (l *License) String() string {
	if l == nil {
		return "<license not found or detected>"
	}

	return l.Name
}

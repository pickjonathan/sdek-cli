package factory

import "fmt"

// ErrUnknownScheme is returned when trying to create a provider for an unregistered URL scheme.
type ErrUnknownScheme struct {
	Scheme string
}

func (e *ErrUnknownScheme) Error() string {
	return fmt.Sprintf("unknown provider scheme: %s (registered schemes: %v)", e.Scheme, ListRegisteredSchemes())
}

// ErrInvalidURL is returned when the provider URL cannot be parsed.
type ErrInvalidURL struct {
	URL    string
	Reason string
}

func (e *ErrInvalidURL) Error() string {
	return fmt.Sprintf("invalid provider URL %q: %s", e.URL, e.Reason)
}

// ValidateProviderURL validates a provider URL without creating a provider.
// Returns nil if the URL is valid and scheme is registered.
func ValidateProviderURL(providerURL string) error {
	scheme, _, err := parseProviderURL(providerURL)
	if err != nil {
		return err
	}

	if !IsSchemeRegistered(scheme) {
		return &ErrUnknownScheme{Scheme: scheme}
	}

	return nil
}

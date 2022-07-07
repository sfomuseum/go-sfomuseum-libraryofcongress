package libraryofcongress

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"net/url"
)

// type Lookup provides an interface for indexing and search Library of Congress (LoC) identifiers.
type Lookup interface {
	// Find() searches for a given LoC identifier.
	Find(context.Context, string) ([]interface{}, error)
	// Append() indexes a LoC record with one or more identifiers.
	Append(context.Context, interface{}) error
}

// type LookupInitializeFunc is a function used to initialize an implementation of the `Lookup` interface.
type LookupInitializationFunc func(ctx context.Context, uri string) (Lookup, error)

// lookup_roster is a `aaronland/go-roster.Roster` instance used to maintain a list of registered `LookupInitializeFunc` initialization functions.
var lookup_roster roster.Roster

// RegisterLookup() associates 'scheme' with 'init_func' in an internal list of avilable `Lookup` implementations.
func RegisterLookup(ctx context.Context, scheme string, init_func LookupInitializationFunc) error {

	err := ensureLookupRoster()

	if err != nil {
		return fmt.Errorf("Failed to ensure roster, %w", err)
	}

	return lookup_roster.Register(ctx, scheme, init_func)
}

// NewLookup() returns a new `Lookup` instance derived from 'uri'. The semantics of and requirements for
// 'uri' as specific to the package implementing the interface.
func NewLookup(ctx context.Context, uri string) (Lookup, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	scheme := u.Scheme

	i, err := lookup_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, fmt.Errorf("Failed to load driver for %s, %w", scheme, err)
	}

	init_func := i.(LookupInitializationFunc)
	return init_func(ctx, uri)
}

// ensureLookupRoster() ensures that a `aaronland/go-roster.Roster` instance used to maintain a list of registered `LookupInitializeFunc`
// initialization functions is present
func ensureLookupRoster() error {

	if lookup_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return fmt.Errorf("Failed to create new roster, %w", err)
		}

		lookup_roster = r
	}

	return nil
}

# go-sfomuseum-libraryofcongress

Go package for working with Library of Congress data in an SFO Museum context.

## Documentation

Documentation is incomplete at this time. For working example code have a look at:

* [cmd/lookup/main.go](cmd/lookup/main.go)
* [lcsh/lookup_test.go](lcsh/lookup_test.go)
* [lcnaf/lookup_test.go](lcnaf/lookup_test.go)

## Tools

### lookup

For example:

```
$> ./bin/lookup -lookup-uri lcsh:// Airplanes
sh85002782 Airplanes

$> ./bin/lookup -lookup-uri lcnaf:// "Lindbergh, Charles A. (Charles Augustus), 1902-1974"
n79100565 Lindbergh, Charles A. (Charles Augustus), 1902-1974
```

## A note about "lookups"

Please have a look at the [A note about "lookup" documentation](https://github.com/sfomuseum/go-sfomuseum-airfield#a-note-about-lookups) in the `go-sfomuseum-airfield` package. The issues outlined there are the same here. The "tl;dr" is:

> It's not great. It's just what we're doing today. The goal right now is to expect a certain amount of "rinse and repeat" in the short term while aiming to make each cycle shorter than the last.

## A note about the data

The data files in this package, and in particular the `data/lcnaf.csv.bz2` file, are very big. As of this writing the data are loaded in to an in-memory `sync.Map` instance which means that a) it takes a non-zero amount of time to load b) consumes a non-trivial amount of memory. As such the `lcnaf` lookup table, derived from data which has 11M rows, only stores label -> ID pointers. It is not possible, at this time, to lookup the label for a given `lcnaf` identifier.

Remember: This package is tailored to SFO Museum's specific needs, and it's specific trade-offs, at the time of writing. As mentioned above "It's not great. It's just what we're doing today."

## See also

* https://github.com/sfomuseum/go-libraryofcongress
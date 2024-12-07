# geosite parser

a parser for [v2fly/domain-list-community](https://github.com/v2fly/domain-list-community) binary release

## usage

```go
import "github.com/fa1se/dlc-parser/geosite"

// read all bytes into memory
// suitable for a modest file size of ~1MB
data, _ := ioutil.ReadFile("dlc.dat")

// returns a map from country code to a list of domains
// or nil on error
sites := geosite.ParseCollection(data)
if sites == nil {
	// probably malformed buffer
}
// directly access the map
selected := sites["google"]
// or use select method 
selected := sites.Select("google")
// attributes are supported in latter way
selected := sites.Select("google@ads")

for _, record := range selected {
	record.Value // string: e.g. "google.com"
	record.Type  // enum: geosite.RECORD_{KEYWORD,REGEXP,DOMAIN,FULL}
	if record.Attr != nil {
		// do something with attributes, see more in implementation
	}
}
```

## test

both `ParseCollection()` and `Select()` can be tested with latest binary release against implementation ported from [v2fly/v2ray-core](https://github.com/v2fly/v2ray-core).

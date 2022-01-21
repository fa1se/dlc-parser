# geosite parser

a parser for [v2fly/domain-list-community](https://github.com/v2fly/domain-list-community) binary release

# usage

```go
import dlc "github.com/fa1se/dlc-parser"
// ...
data, _ := ioutil.ReadFile("dlc.dat")
sites := dlc.ParseCollection(data)
if sites == nil {
	// probably malformed buffer
}
selected := sites.Select("google@cn")
```

# test

both functions can be tested by latest binary release against an implementation ported from [v2fly/v2ray-core](https://github.com/v2fly/v2ray-core).
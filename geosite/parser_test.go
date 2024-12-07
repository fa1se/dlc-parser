package geosite_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/fa1se/dlc-parser/geosite"
	"github.com/stretchr/testify/assert"
	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"

	"github.com/v2fly/v2ray-core/v5/infra/conf/geodata"
	"google.golang.org/protobuf/proto"
)

func LoadBinaryRelease(t *testing.T) []byte {
	const CACHE_FILE = "dlc.dat"
	const DLC_RELEASE_LATEST = "https://github.com/v2fly/domain-list-community/releases/latest/download/dlc.dat"
	cache, err0 := ioutil.ReadFile(CACHE_FILE)
	if err0 == nil {
		return cache
	} else {
		resp, err1 := http.Get(DLC_RELEASE_LATEST)
		assert.Nil(t, err1) // not exactly part of the test
		payload, err2 := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err2) // not exactly part of the test
		ioutil.WriteFile(CACHE_FILE, payload, 0644)
		return payload
	}
}

func GetTestCollection(t *testing.T) geosite.Collection {
	data := LoadBinaryRelease(t)
	// parse dlc.dat
	col := geosite.ParseCollection(data)
	assert.NotNil(t, col)
	return col
}

func GetReferenceParsedList(t *testing.T) routercommon.GeoSiteList {
	data := LoadBinaryRelease(t)
	// reference implementation
	var listRef routercommon.GeoSiteList // load the v2ray way
	err := proto.Unmarshal(data, &listRef)
	assert.Nil(t, err) // not exactly part of the test
	return listRef
}

func GetReferenceSelectedList(t *testing.T, expr string) []*routercommon.Domain {
	loadGeositeWithAttr := func(siteWithAttr string) ([]*routercommon.Domain, error) {
		parts := strings.Split(siteWithAttr, "@")
		if len(parts) == 0 {
			return nil, errors.New("empty rule")
		}
		list := strings.TrimSpace(parts[0])
		attrVal := parts[1:]

		if len(list) == 0 {
			return nil, errors.New("empty listname in rule: " + siteWithAttr)
		}

		loadSite := func(list string) ([]*routercommon.Domain, error) {
			geositeList := GetReferenceParsedList(t)
			for _, site := range geositeList.Entry {
				if strings.EqualFold(site.CountryCode, list) {
					return site.Domain, nil
				}
			}
			return nil, errors.New("list not found: " + list)
		}

		domains, err := loadSite(list)
		if err != nil {
			return nil, err
		}
		type AttributeList struct {
			matcher []geodata.AttributeMatcher
		}
		parseAttrs := func(attrs []string) *AttributeList {
			al := new(AttributeList)
			for _, attr := range attrs {
				trimmedAttr := strings.ToLower(strings.TrimSpace(attr))
				if len(trimmedAttr) == 0 {
					continue
				}
				al.matcher = append(al.matcher, geodata.BooleanMatcher(trimmedAttr))
			}
			return al
		}

		attrs := parseAttrs(attrVal)
		if len(attrs.matcher) == 0 {
			if strings.Contains(siteWithAttr, "@") {
				return nil, errors.New("empty attribute list: " + siteWithAttr)
			}
			return domains, nil
		}
		filteredDomains := make([]*routercommon.Domain, 0, len(domains))
		hasAttrMatched := false
		Match := func(al *AttributeList, domain *routercommon.Domain) bool {
			for _, matcher := range al.matcher {
				if !matcher.Match(domain) {
					return false
				}
			}
			return true
		}
		for _, domain := range domains {
			if Match(attrs, domain) {
				hasAttrMatched = true
				filteredDomains = append(filteredDomains, domain)
			}
		}
		if !hasAttrMatched {
			return nil, errors.New("attribute match no rule: geosite:" + siteWithAttr)
		}
		return filteredDomains, nil
	}
	selected, err := loadGeositeWithAttr(expr)
	assert.Nil(t, err)
	return selected
}

func assertSameDomainList(t *testing.T, expected []*routercommon.Domain, actual geosite.RecordList) {
	t.Logf("size: %d\n", len(expected))
	assert.Equal(t, len(expected), len(actual))
	// sort before side-by-side comparison
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].Value < expected[j].Value
	})
	sort.Slice(actual, func(i, j int) bool {
		return actual[i].Value < actual[j].Value
	})
	for index, expecedDomain := range expected {
		actualDomain := actual[index]
		assertSameDomain(t, expecedDomain, actualDomain)
	}
}

func assertSameDomain(t *testing.T, expected *routercommon.Domain, actual *geosite.Record) {
	t.Logf("\tdomain: '%s', type: %d\n", expected.Value, expected.Type)
	assert.Equal(t, expected.Value, actual.Value)
	assert.Equal(t, int(expected.Type), int(actual.Type))
	assert.Equal(t, len(expected.Attribute), len(actual.Attr))
	for _, expectedAttr := range expected.Attribute { // always singleton
		expectedKey := expectedAttr.Key
		actualValue, actualExist := actual.Attr[expectedKey]
		assert.True(t, actualExist)
		// same attr type
		if expectedValue, ok := expectedAttr.GetTypedValue().(*routercommon.Domain_Attribute_BoolValue); ok {
			t.Logf("\t\tattr: '%s', bool: %t", expectedKey, expectedValue.BoolValue)
			assert.Nil(t, actualValue)
		}
		if value, ok := expectedAttr.GetTypedValue().(*routercommon.Domain_Attribute_IntValue); ok {
			t.Logf("\t\tattr: '%s', int: %d", expectedKey, value.IntValue)
			assert.NotNil(t, actualValue)
			assert.Equal(t, value.IntValue, *actualValue) // never used
		}

	}
}

func TestParseCollection(t *testing.T) {
	col := GetTestCollection(t)
	ref := GetReferenceParsedList(t)
	// same number of entries
	assert.Equal(t, len(ref.Entry), len(col))
	for _, entryRef := range ref.Entry {
		countryCode := strings.ToLower(entryRef.CountryCode)
		t.Logf("entry: '%s'\n", countryCode)
		expectedList := entryRef.Domain
		actualList := col[countryCode]
		assertSameDomainList(t, expectedList, actualList)
	}
}

func TestSelector(t *testing.T) {
	col := GetTestCollection(t)
	exprs := [...]string{
		"category-ads",
		"category-ads-all",
		"tld-!cn",
		"geolocation-cn",
		"geolocation-!cn",
		"cn",
		"apple",
		"google",
		"microsoft",
		"facebook",
		"twitter",
		"telegram",
		"icloud",
		"netflix",
		"apple@cn",
		"google@cn",
		"google@ads",
		"alibaba@ads",
		"baidu@ads",
		"steam",
		"steam@cn",
		"tencent",
		"tencent@ads",
	}
	for _, expr := range exprs {
		actualList := col.Select(expr)
		expectedList := GetReferenceSelectedList(t, expr)
		t.Logf("expr: '%s'\n", expr)
		assertSameDomainList(t, expectedList, actualList)
	}
}

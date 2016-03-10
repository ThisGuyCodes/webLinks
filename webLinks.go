package webLinks

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseLink parses a "Link" header. This accepts only the value portion of
// the header, not the whole header.
func ParseLink(link string) []Link {
	// Strip whitespace
	link = strings.Trim(link, " ")

	thisLink := Link{}
	uriEnd := strings.IndexRune(link, '>')

	thisLink.URI = link[1:uriEnd]

	paramsStart := strings.IndexRune(link[uriEnd:], ';') + uriEnd + 1
	params, paramsEnd := parseLinkParams(link[paramsStart:])
	paramsEnd += paramsStart
	thisLink.Params = params
	nextLink := strings.IndexRune(link[paramsEnd:], ',') + paramsEnd + 1
	if nextLink == paramsEnd {
		return []Link{thisLink}
	}
	return append([]Link{thisLink}, ParseLink(link[nextLink:])...)
}

func parseLinkParams(params string) (map[string]Param, int) {
	paramsEnd := strings.IndexRune(params, ',')
	if paramsEnd == -1 {
		paramsEnd = len(params)
	}
	pStrs := strings.Split(params[:paramsEnd], ";")
	mapped := make(map[string]Param, len(pStrs))
	for _, p := range pStrs {
		key, value := parseParam(p)
		mapped[key] = value
	}
	return mapped, paramsEnd
}

func parseParam(param string) (string, Param) {
	// Trim whitespace
	param = strings.Trim(param, " ")
	parts := strings.SplitN(param, "=", 2)
	if len(parts) != 2 {
		// This does not fall within the spec, so 'best effort'
		return param, Param{}
	}
	key, value := parts[0], parts[1]

	enc := "us-ascii"
	lang := "en-us"

	if strings.HasSuffix(key, "*") {
		// value is URL encoded and *may* contain encoding+language meta

		// Strip the * indicator
		key = key[:len(key)-1]

		// Split out the encoding information
		valueParts := strings.Split(value, "'")
		if len(valueParts) == 3 {
			enc = valueParts[0]
			lang = valueParts[1]
			value = valueParts[2]
		}
		// It's just encoded, leave the defaults

		// Decode this sucker
		decoded, err := url.QueryUnescape(value)
		if err == nil {
			value = decoded
		}
		// not within spec, just leave it encoded
	} else {
		// It's not encoded, but it's quoted
		// Let's dequote it
		var dequoted string
		n, err := fmt.Sscanf(value, "%q", &dequoted)
		if n == 1 && err == nil {
			value = dequoted
		}
		// ???, just leave it as is
	}
	p := Param{
		Value: value,
		Enc:   enc,
		Lang:  lang,
	}
	return key, p
}

// Link represents a link from a parsed Link header
type Link struct {
	URI    string
	Params map[string]Param
}

// Param represents a single link parameter
// This is necessary because parameters can state their own encoding, and be
// multipart. See http://tools.ietf.org/html/rfc2231
//
// Although the encoding may not be UTF-8 compliant, we still return a UTF-8
// string. Other encodings must be handled by the caller if desired.
type Param struct {
	Value string
	Enc   string
	Lang  string
}

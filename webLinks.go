package webLinks

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
)

// ParseLink parses a "Link" header. This accepts only the value portion of
// the header, not the whole header.
func ParseLink(link string) []Link {
	// Bytes!
	bLink := []byte(link)
	// Strip whitespace
	bLink = bytes.Trim(bLink, " ")

	thisLink := Link{}
	uriEnd := bytes.IndexRune(bLink, '>')

	thisLink.URI = string(bLink[1:uriEnd])

	paramsStart := strings.IndexRune(link[uriEnd:], ';') + uriEnd + 1
	params, paramsEnd := parseLinkParams(bLink[paramsStart:])
	paramsEnd += paramsStart
	thisLink.Params = params
	nextLink := strings.IndexRune(link[paramsEnd:], ',') + paramsEnd + 1
	if nextLink == paramsEnd {
		return []Link{thisLink}
	}
	return append([]Link{thisLink}, ParseLink(link[nextLink:])...)
}

func parseLinkParams(params []byte) (map[string]Param, int) {
	paramsEnd := bytes.IndexRune(params, ',')
	if paramsEnd == -1 {
		paramsEnd = len(params)
	}
	pStrs := bytes.Split(params[:paramsEnd], []byte(";"))
	mapped := make(map[string]Param, len(pStrs))
	for _, p := range pStrs {
		key, value := parseParam(p)
		mapped[key] = value
	}
	return mapped, paramsEnd
}

func parseParam(param []byte) (string, Param) {
	thisParam := Param{
		Enc:  "us-ascii",
		Lang: "en-us",
	}

	// Trim whitespace
	param = bytes.Trim(param, " ")
	parts := bytes.SplitN(param, []byte("="), 2)
	if len(parts) != 2 {
		// This does not fall within the spec, so 'best effort'
		return string(param), thisParam
	}
	key, value := parts[0], parts[1]

	if bytes.HasSuffix(key, []byte("*")) {
		// value is URL encoded and *may* contain encoding+language meta

		// Strip the * indicator
		key = key[:len(key)-1]

		// Split out the encoding information
		valueParts := bytes.Split(value, []byte("'"))
		if len(valueParts) == 3 {
			thisParam.Enc = string(valueParts[0])
			thisParam.Lang = string(valueParts[1])
			value = valueParts[2]
		}
		// It's just encoded, leave the defaults

		// Decode this sucker
		sValue := string(value)
		decoded, err := url.QueryUnescape(sValue)
		if err == nil {
			sValue = decoded
		}
		// not within spec, just leave it encoded
		thisParam.Value = sValue
	} else {
		// It's not encoded, but it's quoted
		// Let's dequote it
		var dequoted string
		sValue := string(value)
		n, err := fmt.Sscanf(sValue, "%q", &dequoted)
		if n == 1 && err == nil {
			thisParam.Value = dequoted
		} else {
			// ???, just leave it as is
			thisParam.Value = sValue
		}
	}
	return string(key), thisParam
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

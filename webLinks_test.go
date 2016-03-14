package webLinks_test

import (
	"testing"

	"github.com/conslo/webLinks"
)

var tests = []struct {
	input string
	links []webLinks.Link
}{
	{
		`<http://example.com/TheBook/chapter2>; rel="previous"; title="previous chapter"`,
		[]webLinks.Link{
			{
				"http://example.com/TheBook/chapter2",
				map[string]webLinks.Param{
					"rel":   {Value: "previous", Enc: "us-ascii", Lang: "en-us"},
					"title": {Value: "previous chapter", Enc: "us-ascii", Lang: "en-us"},
				},
			},
		},
	},
	{
		`</>; rel="http://example.net/foo"`,
		[]webLinks.Link{
			{
				"/",
				map[string]webLinks.Param{
					"rel": {Value: "http://example.net/foo", Enc: "us-ascii", Lang: "en-us"},
				},
			},
		},
	},
	{
		`</TheBook/chapter2>; rel="previous"; title*=UTF-8'de'letztes%20Kapitel, </TheBook/chapter4>; rel="next"; title*=UTF-8'de'n%c3%a4chstes%20Kapitel`,
		[]webLinks.Link{
			{
				"/TheBook/chapter2",
				map[string]webLinks.Param{
					"rel":   {Value: "previous", Enc: "us-ascii", Lang: "en-us"},
					"title": {Value: "letztes Kapitel", Enc: "UTF-8", Lang: "de"},
				},
			},
			{
				"/TheBook/chapter4",
				map[string]webLinks.Param{
					"rel":   {Value: "next", Enc: "us-ascii", Lang: "en-us"},
					"title": {Value: "nächstes Kapitel", Enc: "UTF-8", Lang: "de"},
				},
			},
		},
	},
}

func TestParseLinksURI(t *testing.T) {
	t.Parallel()
	for _, test := range tests {
		links := webLinks.Parse(test.input)
		if len(links) != len(test.links) {
			t.Fatalf("Length mismatch, got %d expected %d\n", len(links), len(test.links))
		}
		for i, link := range links {
			if link.URI != test.links[i].URI {
				t.Fatalf("Got the wrong URI, got %q expected %q\n", link.URI, test.links[i].URI)
			}
		}
	}
}

func TestParseLinksParamValue(t *testing.T) {
	t.Parallel()
	for _, test := range tests {
		links := webLinks.Parse(test.input)
		if len(links) != len(test.links) {
			t.Fatalf("Length mismatch, got %d expected %d\n", len(links), len(test.links))
		}
		for i, link := range links {
			if len(link.Params) != len(test.links[i].Params) {
				t.Fatalf("Length mismatch, got %d expected %d\n", len(link.Params), len(test.links[i].Params))
			}
			for k, v := range test.links[i].Params {
				if link.Params[k] != v {
					t.Fatalf("Value mismatch, got %q expected %q\n", link.Params[k], v)
				}
			}
		}
	}
}

func TestParseLinksIntoMap(t *testing.T) {
	t.Parallel()
	links := webLinks.Links{
		webLinks.Link{
			URI: "some uri",
			Params: map[string]webLinks.Param{
				"rel": {
					Value: "some relation",
					Enc:   "doesn't matter",
					Lang:  "this either",
				},
			},
		},
		webLinks.Link{
			URI: "another uri",
			Params: map[string]webLinks.Param{
				"rel": {
					Value: "another relation",
				},
			},
		},
	}

	these := links.Map()

	if these["some relation"].URI != "some uri" {
		t.Fatalf("Got bad relation in map. Got %q expected %1\n", these["some relation"].URI, "some uri")
	}

	if these["another relation"].URI != "another uri" {
		t.Fatalf("Got bad relation in map. Got %q expected %1\n", these["another relation"].URI, "another uri")
	}
}

func BenchmarkParseLinksFancy(b *testing.B) {
	this := `</TheBook/chapter2>; rel="previous"; title*=UTF-8'de'letztes%20Kapitel, </TheBook/chapter4>; rel="next"; title*=UTF-8'de'n%c3%a4chstes%20Kapitel`
	b.SetBytes(int64(len([]byte(this))))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		webLinks.Parse(this)
	}
}

func BenchmarkParseLinksSimple(b *testing.B) {
	this := `<http://example.com/TheBook/chapter2>; rel="previous"; title="previous chapter"`
	b.SetBytes(int64(len([]byte(this))))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		webLinks.Parse(this)
	}
}

func BenchmarkParseLinksSimplist(b *testing.B) {
	this := `<http://example.com/>; rel="previous"`
	b.SetBytes(int64(len([]byte(this))))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		webLinks.Parse(this)
	}
}

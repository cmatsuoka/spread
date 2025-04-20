package spread_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/snapcore/spread/spread"

	. "gopkg.in/check.v1"
)

type googleSuite struct{}

var _ = Suite(&googleSuite{})

func (s *googleSuite) TestImagesCache(c *C) {
	n := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch n {
		case 0:
			c.Check(r.URL.Path, Equals, "/compute/v1/projects/snapd/global/images")
			w.Write([]byte(`
                {
                    "items": [{"status":"READY","name":"ubuntu-1910-64-v20190826", "description":"ubuntu-19.10-64"}, {"status":"READY","name":"ubuntu-1904-64-v20190726", "description":"ubuntu-19.04-64"}]
                }
                `))
		case 1:
			c.Check(r.URL.Path, Equals, "/compute/v1/projects/other-project/global/images")
			w.Write([]byte(`
                {
                    "items": [{"status":"READY","name":"ubuntu-2004-64-v20190826", "description":"ubuntu-19.10-64"}]
                }
                `))
		default:
			c.Errorf("unexpected number of requests")
		}
		n++
	}))
	defer mockServer.Close()

	g := spread.FakeGoogleProvider(mockServer.URL, nil, nil, nil)
	c.Assert(g, NotNil)

	// Request the project images
	images, err := spread.ProjectImages(g, "snapd")
	c.Assert(err, IsNil)
	c.Assert(images, HasLen, 2)
	c.Check(images, DeepEquals, []spread.GoogleImage{
		{"snapd", "ubuntu-1910-64-v20190826", "", []string{"ubuntu", "19.10", "64"}},
		{"snapd", "ubuntu-1904-64-v20190726", "", []string{"ubuntu", "19.04", "64"}},
	})
	c.Check(n, Equals, 1)

	// do it again, now it comes from the cache
	images, err = spread.ProjectImages(g, "snapd")
	c.Assert(err, IsNil)
	c.Assert(images, HasLen, 2)
	c.Check(images, DeepEquals, []spread.GoogleImage{
		{"snapd", "ubuntu-1910-64-v20190826", "", []string{"ubuntu", "19.10", "64"}},
		{"snapd", "ubuntu-1904-64-v20190726", "", []string{"ubuntu", "19.04", "64"}},
	})
	c.Check(n, Equals, 1)

	// again, this time for another project
	images, err = spread.ProjectImages(g, "other-project")
	c.Assert(err, IsNil)
	c.Assert(images, HasLen, 1)
	c.Check(n, Equals, 2)
}

type ensureLabelFormatTest struct {
	label          string
	formattedLabel string
}

var ensureLabelFormatTests = []ensureLabelFormatTest{{
	label:          "",
	formattedLabel: "",
}, {
	label:          "user123",
	formattedLabel: "user123",
}, {
	label:          "User.Name",
	formattedLabel: "user_name",
}, {
	label:          "user.name@my-domain.com",
	formattedLabel: "user_name_my-domain_com",
}, {
	label:          "123456789.123456789@123456789.123456789.123456789.123456789.1234567890",
	formattedLabel: "123456789_123456789_123456789_123456789_123456789_123456789_123",
}}

func (s *googleSuite) TestEnsureLabelFormat(c *C) {
	for _, tc := range ensureLabelFormatTests {
		repl := spread.EnsureGoogleLabelFormat(tc.label)
		c.Check(repl, Equals, tc.formattedLabel)
	}
}

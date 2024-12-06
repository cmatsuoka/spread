package spread_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/snapcore/spread/spread"

	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

func Test(t *testing.T) { TestingT(t) }

type FilterSuite struct{}

var _ = Suite(&FilterSuite{})

func (s *FilterSuite) TestFilter(c *C) {
	job := &spread.Job{Name: "backend:image:suite/test:variant"}

	pass := []string{
		"backend",
		"backend:",
		"image",
		":image:",
		"suite/test",
		"suit...est",
		"suite/",
		"/test",
		":variant",
		"...",
		"im...",
		"...ge",
	}

	block := []string{
		"nothing",
		"noth...",
		"...hing",
		":backend",
		"suite",
		"test",
	}

	for _, s := range pass {
		f, err := spread.NewFilter([]string{s})
		c.Assert(err, IsNil)
		c.Assert(f.Pass(job), Equals, true, Commentf("Filter: %q", s))
	}

	for _, s := range block {
		f, err := spread.NewFilter([]string{s})
		c.Assert(err, IsNil)
		c.Assert(f.Pass(job), Equals, false, Commentf("Filter: %q", s))
	}
}

type projectSuite struct{}

var _ = Suite(&projectSuite{})

func (s *projectSuite) TestLoad(c *C) {
	spreadYaml := []byte(`project: mock-prj
path: /some/path
backends:
 google:
  key: some-key
  plan: global-plan
  systems:
   - system-1:
   - system-2:
      plan: plan-for-2
   - system-3:
suites:
 tests/:
  summary: mock tests
`)

	for _, tc := range []struct {
		name        string // the file name to write the project to
		prjFilename string // the file name specified by the user
		errMsg      string // the expected error message
	}{
		{"spread.yaml", "", ""},
		{".spread.yaml", "", ""},
		{"other.yaml", "", "cannot load project from .*: cannot find spread.yaml or .spread.yaml"},
		{"custom.yaml", "custom.yaml", ""},
		{"subdir/custom.yaml", "subdir/custom.yaml", ""},
		{"subdir/custom.yaml", "subdir/.././subdir/custom.yaml", ""},
		{"spread.yaml", "/custom.yaml", "cannot load project: open /custom.yaml: no such file or directory"},
		{"spread.yaml", "custom.yaml", "cannot load project: open /.*/custom.yaml: no such file or directory"},
	} {
		tmpdir := c.MkDir()
		err := os.MkdirAll(filepath.Join(tmpdir, "subdir"), 0755)
		c.Assert(err, IsNil)

		err = os.WriteFile(filepath.Join(tmpdir, tc.name), spreadYaml, 0644)
		c.Assert(err, IsNil)
		err = os.MkdirAll(filepath.Join(tmpdir, "tests"), 0755)
		c.Assert(err, IsNil)

		if tc.prjFilename != "" {
			os.Setenv("SPREAD_PROJECT_FILE", tc.prjFilename)
		}
		defer os.Unsetenv("SPREAD_PROJECT_FILE")

		prj, err := spread.Load(tmpdir)
		if tc.errMsg == "" {
			c.Assert(err, IsNil, Commentf("test case: %+v", tc))
			backend := prj.Backends["google"]
			c.Check(backend.Name, Equals, "google")
			c.Check(backend.Systems["system-1"].Plan, Equals, "global-plan")
			c.Check(backend.Systems["system-2"].Plan, Equals, "plan-for-2")
			c.Check(backend.Systems["system-3"].Plan, Equals, "global-plan")
		} else {
			c.Assert(err, ErrorMatches, tc.errMsg, Commentf("test case: %+v", tc))
		}
	}
}

func (s *projectSuite) TestOptionalInt(c *C) {
	optInts := struct {
		Priority spread.OptionalInt `yaml:"priority"`
		NotSet   spread.OptionalInt `yaml:"not-set"`
	}{}
	inp := []byte("priority: 100")

	err := yaml.Unmarshal(inp, &optInts)
	c.Assert(err, IsNil)
	c.Check(optInts.Priority.IsSet, Equals, true)
	c.Check(optInts.Priority.Value, Equals, int64(100))
	c.Check(optInts.Priority.String(), Equals, "100")

	c.Check(optInts.NotSet.IsSet, Equals, false)
	c.Check(optInts.NotSet.Value, Equals, int64(0))
	c.Check(optInts.NotSet.String(), Equals, "0")
}

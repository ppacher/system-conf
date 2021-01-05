package conf

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDropInSearchPaths(t *testing.T) {
	paths := DropInSearchPaths("foo-bar-baz.task", "/lib/")

	assert.Equal(t, []string{
		"/lib/task.d",
		"/lib/foo-.task.d",
		"/lib/foo-bar-.task.d",
		"/lib/foo-bar-baz.task.d",
	}, paths)
}

type fakeFileInfo struct {
	name  string
	isDir bool
}

func (t *fakeFileInfo) IsDir() bool {
	return t.isDir
}
func (t *fakeFileInfo) Name() string {
	return t.name
}
func (*fakeFileInfo) ModTime() time.Time { return time.Now() }
func (*fakeFileInfo) Mode() os.FileMode  { return 0600 }
func (*fakeFileInfo) Size() int64        { return 100 }
func (*fakeFileInfo) Sys() interface{}   { return nil }

func fakeFile(name string, dir bool) os.FileInfo {
	return &fakeFileInfo{name, dir}
}

func TestSearchDropinFiles(t *testing.T) {
	// restore readDir after this test case
	defer func() {
		readDir = ioutil.ReadDir
	}()
	readDir = func(path string) ([]os.FileInfo, error) {
		switch {
		case strings.HasPrefix(path, "/lib/task.d"):
			return []os.FileInfo{
				fakeFile("test", false),
				fakeFile("dir.conf", true),
				fakeFile("10-overwrite.conf", false),
				fakeFile("20-task.d.conf", false),
			}, nil
		case strings.HasPrefix(path, "/lib/foo-.task.d"):
			return []os.FileInfo{
				fakeFile("test2", false),
				fakeFile("10-overwrite.conf", false),
				fakeFile("30-foo-task.d.conf", false),
			}, nil
		case strings.HasPrefix(path, "/lib/foo-bar-baz.task.d"):
			return []os.FileInfo{
				fakeFile("10-overwrite.conf", false),
			}, nil
		}

		return nil, os.ErrNotExist
	}

	paths, err := SearchDropinFiles("foo-bar-baz.task", []string{"/lib/"})
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"/lib/foo-bar-baz.task.d/10-overwrite.conf",
		"/lib/task.d/20-task.d.conf",
		"/lib/foo-.task.d/30-foo-task.d.conf",
	}, paths)
}

func TestApplyDropIns(t *testing.T) {
	specs := FileSpec{
		"test": SectionSpec{
			{
				Type: StringType,
				Name: "Single",
			},
			{
				Name: "Slice1",
				Type: StringSliceType,
			},
			{
				Name: "Slice2",
				Type: StringSliceType,
			},
		},
	}
	tsk := &File{
		Sections: []Section{
			{
				Name: "Test",
				Options: Options{
					{
						Name:  "Single",
						Value: "TSK",
					},
					{
						Name:  "Slice1",
						Value: "TSK",
					},
				},
			},
		},
	}
	d1 := &DropIn{
		Sections: []Section{
			{
				Name: "Test",
				Options: Options{
					{
						Name:  "Single",
						Value: "d1",
					},
					{
						Name:  "Slice2",
						Value: "d1",
					},
					{
						Name:  "Slice2",
						Value: "d1",
					},
					{
						Name:  "Slice1",
						Value: "d1",
					},
				},
			},
		},
	}

	d2 := &DropIn{
		Sections: []Section{
			{
				Name: "Test",
				Options: Options{
					{
						Name:  "Slice1",
						Value: "", // clear all values
					},
					{
						Name:  "Slice1",
						Value: "d2",
					},
				},
			},
		},
	}

	res := tsk.Clone()
	err := ApplyDropIns(res, []*DropIn{d1, d2}, specs)
	assert.NoError(t, err)
	assert.Equal(t, &File{
		Sections: []Section{
			{
				Name: "Test",
				Options: Options{
					{
						Name:  "Single",
						Value: "d1",
					},
					{
						Name:  "Slice2",
						Value: "d1",
					},
					{
						Name:  "Slice2",
						Value: "d1",
					},
					{
						Name:  "Slice1",
						Value: "d2",
					},
				},
			},
		},
	}, res)
}

func TestApplyDropInsNotAllowed(t *testing.T) {
	tsk := &File{
		Sections: []Section{
			{
				Name: "Test",
			},
			{
				Name: "Test",
			},
		},
	}

	err := ApplyDropIns(tsk,
		[]*DropIn{
			{
				Sections: []Section{
					{ // section Test is not allowed because it's not unique in tsk
						Name: "Test",
					},
				},
			},
		},
		FileSpec{
			"test": nil,
		},
	)

	assert.Error(t, err)
}

func TestApplyDropInsSectionNotExists(t *testing.T) {
	tsk := &File{
		Sections: []Section{},
	}

	err := ApplyDropIns(tsk,
		[]*DropIn{
			{
				Sections: []Section{
					{ // section Test is not allowed because it's not unique in tsk
						Name: "Unknown",
					},
				},
			},
		},
		FileSpec{
			"test": nil,
		},
	)

	assert.Error(t, err)
}

func TestApplyDropInsOptionNotExists(t *testing.T) {
	tsk := &File{
		Sections: []Section{
			{
				Name: "Test",
			},
		},
	}

	err := ApplyDropIns(tsk,
		[]*DropIn{
			{
				Sections: []Section{
					{ // section Test is not allowed because it's not unique in tsk
						Name: "Test",
						Options: Options{
							{
								Name: "does-not-exist",
							},
						},
					},
				},
			},
		},
		FileSpec{
			"test": nil,
		},
	)

	assert.Error(t, err)
}

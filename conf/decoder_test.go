package conf_test

import (
	"testing"
	"time"

	"github.com/ppacher/system-conf/conf"
	"github.com/stretchr/testify/assert"
)

func TestFileSpecDecode(t *testing.T) {
	spec := conf.FileSpec{
		"Global": conf.SectionSpec{
			{
				Name: "LogLevel",
				Type: conf.StringType,
			},
			{
				Name: "Fields",
				Type: conf.StringSliceType,
			},
		},
		"LogFile": conf.SectionSpec{
			{
				Name: "Path",
				Type: conf.StringType,
			},
			{
				Name: "Rotate",
				Type: conf.BoolType,
			},
			{
				Name: "MaxAge",
				Type: conf.DurationType,
			},
		},
	}

	type TestGlobal struct {
		LogLevel string
		Fields   []string
	}

	type TestLogFile struct {
		Path       string
		RotateFile bool `option:"Rotate"`
		MaxAge     time.Duration
	}

	type Test struct {
		Global   TestGlobal
		LogFiles []TestLogFile `section:"LogFile,required"`
	}

	f := &conf.File{
		Sections: []conf.Section{
			{
				Name: "Global",
				Options: conf.Options{
					{
						Name:  "LogLevel",
						Value: "info",
					},
					{
						Name:  "Fields",
						Value: "Hostname",
					},
					{
						Name:  "Fields",
						Value: "Error",
					},
				},
			},
			{
				Name: "LogFile",
				Options: conf.Options{
					{
						Name:  "Path",
						Value: "/var/log/path1",
					},
					{
						Name:  "Rotate",
						Value: "yes",
					},
					{
						Name:  "MaxAge",
						Value: "10h",
					},
				},
			},
			{
				Name: "LogFile",
				Options: conf.Options{
					{
						Name:  "Path",
						Value: "/var/log/path2",
					},
					{
						Name:  "Rotate",
						Value: "no",
					},
					{
						Name:  "MaxAge",
						Value: "24h",
					},
				},
			},
		},
	}

	var target Test
	err := spec.Decode(f, &target)
	assert.NoError(t, err)

	assert.Equal(t, Test{
		Global: TestGlobal{
			LogLevel: "info",
			Fields:   []string{"Hostname", "Error"},
		},
		LogFiles: []TestLogFile{
			{
				Path:       "/var/log/path1",
				RotateFile: true,
				MaxAge:     time.Hour * 10,
			},
			{
				Path:       "/var/log/path2",
				RotateFile: false,
				MaxAge:     time.Hour * 24,
			},
		},
	}, target)
}

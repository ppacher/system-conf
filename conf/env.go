package conf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

func ParseFromEnv(prefix string, env []string, reg SectionRegistry) (*File, error) {
	envFile := new(File)

	sections := make(map[string][]*Section)

	for varName, varValue := range toMap(env) {
		parts := strings.Split(varName, "_")
		if !strings.EqualFold(parts[0], prefix) {
			continue
		}

		if len(parts) < 3 {
			// PREFIX_SECTION_KEY requires at least 3 parts.
			continue
		}
		sectionName := strings.ToLower(parts[1])
		optName := parts[2]
		sectionIdx := 0

		if len(parts) >= 4 {
			idx, err := strconv.ParseInt(parts[2], 10, 0)
			if err == nil {
				sectionIdx = int(idx)
				optName = parts[3]
			}
		}

		optReg, ok := reg.OptionsForSection(sectionName)
		if !ok {
			// Skip unknown section name
			continue
		}

		var sec *Section
		switch {
		case len(sections[sectionName]) == sectionIdx:
			sec = &Section{
				Name: sectionName,
			}
			sections[sectionName] = append(sections[sectionName], sec)

		case len(sections[sectionName]) < sectionIdx:
			return nil, fmt.Errorf("invalid index %d for section %s (in %+v)", sectionIdx, sectionName, sections[sectionName])

		case sectionIdx <= len(sections[sectionName])-1:
			sec = sections[sectionName][sectionIdx]

		default:
			return nil, fmt.Errorf("cannot get section with index %d in %+v", sectionIdx, sections[sectionName])
		}

		optSpec, ok := optReg.GetOption(strings.ToLower(optName))
		if !ok {
			return nil, fmt.Errorf("invalid option name %s for section %s", optName, sectionName)
		}

		var values []string
		if optSpec.Type.IsSliceType() {
			var err error
			values, err = shlex.Split(varValue)
			if err != nil {
				return nil, fmt.Errorf("failed to parse option value for %s.%s: %w", sectionName, optName, err)
			}
		} else {
			values = []string{varValue}
		}

		for _, val := range values {
			// TODO(ppacher): verify option values now or rely on ValidateFile?
			sec.Options = append(sec.Options, Option{
				Name:  optSpec.Name,
				Value: val,
			})
		}
	}

	for _, secs := range sections {
		for _, sec := range secs {
			envFile.Sections = append(envFile.Sections, *sec)
		}
	}

	return envFile, nil
}

func toMap(env []string) map[string]string {
	r := map[string]string{}
	for _, e := range env {
		p := strings.SplitN(e, "=", 2)
		r[p[0]] = p[1]
	}
	return r
}

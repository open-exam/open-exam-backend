package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-exam/open-exam-backend/util"
)

var (
	poolingTypes = []string{"roundRobin", "random"}
	gridTypes = []string{"plugin", "box"}
)

type Exam struct {
	Sections []SectionIntermediate
}

type SectionIntermediate struct {
	Section
}

type SectionSwitch struct {
	SectionType string
}

type Section struct {
	Name string `json:"name"`
	SectionSwitch
	QuestionConfig QuestionConfig `json:"questionConfig"`
	Layout Layout `json:"layout"`
}

type QuestionConfig struct {
	SectionSwitch
	*StandardQuestions
	*PooledQuestions
	*SetQuestions
	*PooledSetQuestions
	*CustomQuestions
}

type StandardQuestions struct {
	Questions []uint64 `json:"questions"`
}

type PooledQuestions struct {
	NumQuestions int `json:"numQuestions"`
	Type string `json:"type"`
	PoolId uint64 `json:"poolId"`
}

type SetQuestions struct {
	Sets []int `json:"sets"`
}

type PooledSetQuestions struct {
	NumQuestions int `json:"numQuestions"`
	Type string `json:"type"`
	PooledSets [][]int `json:"pooledSets"`
}

type CustomQuestions struct {
	PluginId uint64 `json:"pluginId"`
	SystemState bool `json:"systemState"`
	UserState bool `json:"userState"`
}

type Layout struct {
	Grid [][]Grid `json:"grid"`
}

type Grid struct {
	Type string `json:"type"`
	PluginId uint64 `json:"pluginId"`
	Css string `json:"css"`
}

func (res *SectionIntermediate) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &res.SectionSwitch); err != nil {
		return err
	}
	res.QuestionConfig.SectionType = res.SectionType
	return json.Unmarshal(data, &res.Section)
}

func (res *QuestionConfig) UnmarshalJSON(data []byte) error {
	switch res.SectionType {
	case "standard": {
		res.StandardQuestions = &StandardQuestions{}
		return json.Unmarshal(data, res.StandardQuestions)
	}
	case "pooled": {
		res.PooledQuestions = &PooledQuestions{}
		return json.Unmarshal(data, res.PooledQuestions)
	}
	case "set": {
		res.SetQuestions = &SetQuestions{}
		return json.Unmarshal(data, res.SetQuestions)
	}
	case "pooledSet": {
		res.PooledSetQuestions = &PooledSetQuestions{}
		return json.Unmarshal(data, res.PooledSetQuestions)
	}
	case "custom": {
		res.CustomQuestions = &CustomQuestions{}
		return json.Unmarshal(data, res.CustomQuestions)
	}
	default:
		return nil
	}
}

func (res *Exam) Validate() error {
	if len(res.Sections) == 0 {
		return errors.New("exam sections are not specified")
	}

	for i, e := range res.Sections {
		if len(e.Name) == 0 {
			return errors.New(fmt.Sprintf("empty section name at index `%d`", i))
		}

		switch e.SectionType {
		case "standard": {
			if len(e.QuestionConfig.StandardQuestions.Questions) == 0 {
				return errors.New(fmt.Sprintf("no questions specified in section `%s`", e.Name))
			}
		}
		case "pooled": {
			if e.QuestionConfig.PooledQuestions.NumQuestions == 0 {
				return errors.New(fmt.Sprintf("number of questions cannot be zero in section `%s`", e.Name))
			}

			if util.IsInList(e.QuestionConfig.PooledQuestions.Type, &poolingTypes) == -1 {
				return errors.New(fmt.Sprintf("pooling type not specified in section `%s`", e.Name))
			}

			if e.QuestionConfig.PooledQuestions.PoolId == 0 {
				return errors.New(fmt.Sprintf("poolId not specified in section `%s`", e.Name))
			}
		}
		case "pooledSet": {
			if e.QuestionConfig.PooledSetQuestions.NumQuestions == 0 {
				return errors.New(fmt.Sprintf("number of questions cannot be zero in section `%s`", e.Name))
			}

			if util.IsInList(e.QuestionConfig.PooledSetQuestions.Type, &poolingTypes) == -1 {
				return errors.New(fmt.Sprintf("pooling type not specified in section `%s`", e.Name))
			}

			if len(e.QuestionConfig.PooledSetQuestions.PooledSets) == 0 {
				return errors.New(fmt.Sprintf("pooled sets not specified in section `%s`", e.Name))
			}

			for i, _ := range e.QuestionConfig.PooledSetQuestions.PooledSets {
				if len(e.QuestionConfig.PooledSetQuestions.PooledSets[i]) != 2 {
					return errors.New(fmt.Sprintf("pooled set config at [%d] in section `%s` not found", i, e.Name))
				}
			}
		}
		case "set": {
			if len(e.QuestionConfig.SetQuestions.Sets) == 0 {
				return errors.New(fmt.Sprintf("sets not specified in section `%s`", e.Name))
			}
		}
		case "custom": {
			if e.QuestionConfig.CustomQuestions.PluginId == 0 {
				return errors.New(fmt.Sprintf("pluginId not specified in section `%s`", e.Name))
			}
		}
		default: {
			return errors.New(fmt.Sprintf("unknown section type in section `%s`", e.Name))
		}
		}

		if len(e.Layout.Grid) == 0 {
			return errors.New(fmt.Sprintf("empty grid in section `%s`", e.Name))
		}

		for outer, outerE := range e.Layout.Grid {
			for inner, innerE := range outerE {
				if util.IsInList(innerE.Type, &gridTypes) == -1 {
					return errors.New(fmt.Sprintf("unknown grid item type at [%d][%d] in section `%s`", outer, inner, e.Name))
				}

				switch innerE.Type {
				case "plugin": {
					if innerE.PluginId == 0 {
						return errors.New(fmt.Sprintf("pluginId not specified in grid at [%d][%d] in section `%s`", outer, inner, e.Name))
					}
				}
				}
			}
		}
	}

	return nil
}
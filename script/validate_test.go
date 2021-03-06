package script

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudradar-monitoring/tacoscript/tasks"
)

type RequirementsTaskMock struct {
	RequirementsToGive []string
	Path               string
}

func (rtm RequirementsTaskMock) GetName() string {
	return ""
}

func (rtm RequirementsTaskMock) Execute(ctx context.Context) tasks.ExecutionResult {
	return tasks.ExecutionResult{}
}

func (rtm RequirementsTaskMock) Validate() error {
	return nil
}

func (rtm RequirementsTaskMock) GetPath() string {
	return rtm.Path
}

func (rtm RequirementsTaskMock) GetRequirements() []string {
	return rtm.RequirementsToGive
}

type errorExpectation struct {
	messagePrefix  string
	availableParts []string
}

func TestScriptsValidation(t *testing.T) {
	testCases := []struct {
		name             string
		scripts          tasks.Scripts
		errorExpectation errorExpectation
	}{
		{
			name: "cycle of 2 scripts requiring each other",
			scripts: tasks.Scripts{
				{
					ID: "script one",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script two"},
							Path:               "script one.RequirementsTaskMock[0]",
						},
					},
				},
				{
					ID: "script two",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script one"},
							Path:               "script two.RequirementsTaskMock[0]",
						},
					},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "cyclic requirements are detected",
				availableParts: []string{
					"script one",
					"script two",
				},
			},
		},
		{
			name: "no requirement cycle",
			scripts: tasks.Scripts{
				{
					ID: "script 3",
					Tasks: []tasks.Task{
						RequirementsTaskMock{},
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 4"},
						},
					},
				},
				{
					ID:    "script 4",
					Tasks: []tasks.Task{},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "",
			},
		},
		{
			name: "circling cycle",
			scripts: tasks.Scripts{
				{
					ID: "script 6",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 7"},
						},
					},
				},
				{
					ID: "script 7",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 8"},
						},
					},
				},
				{
					ID: "script 8",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 6"},
						},
					},
				},
				{
					ID: "script 9",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 6"},
						},
					},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "cyclic requirements are detected",
				availableParts: []string{
					"script 6",
					"script 7",
					"script 8",
				},
			},
		},
		{
			name: "many_scripts_no_cycle",
			scripts: tasks.Scripts{
				{
					ID: "script 10",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 11"},
						},
					},
				},
				{
					ID: "script 11",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 12", "script 13"},
						},
					},
				},
				{
					ID: "script 13",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{},
						},
					},
				},
				{
					ID: "script 12",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{},
						},
					},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "",
			},
		},
		{
			name: "require itself",
			scripts: tasks.Scripts{
				{
					ID: "script 20",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 20"},
							Path:               "script 20 task 1 path 1",
						},
					},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "task at path 'script 20 task 1 path 1' cannot require own script 'script 20'",
			},
		},
		{
			name: "multiple required scripts not found",
			scripts: tasks.Scripts{
				{
					ID: "script 30",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 31", "script 32"},
							Path:               "path 1",
						},
					},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "missing required scripts",
				availableParts: []string{
					"'script 31' at path 'path 1.require[0]'",
					"'script 32' at path 'path 1.require[1]'",
				},
			},
		},
		{
			name: "all required scripts are found",
			scripts: tasks.Scripts{
				{
					ID: "script 32",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 33", "script 34"},
						},
					},
				},
				{
					ID:    "script 33",
					Tasks: []tasks.Task{RequirementsTaskMock{}},
				},
				{
					ID:    "script 34",
					Tasks: []tasks.Task{RequirementsTaskMock{}},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "",
			},
		},
		{
			name: "some required scripts not found",
			scripts: tasks.Scripts{
				{
					ID: "script 35",
					Tasks: []tasks.Task{
						RequirementsTaskMock{
							RequirementsToGive: []string{"script 36"},
						},
					},
				},
				{
					ID: "script 36",
					Tasks: []tasks.Task{RequirementsTaskMock{
						RequirementsToGive: []string{"script 37"},
						Path:               "path 36",
					}},
				},
				{
					ID: "script 38",
					Tasks: []tasks.Task{RequirementsTaskMock{
						RequirementsToGive: []string{"script 40"},
						Path:               "path 38",
					}},
				},
			},
			errorExpectation: errorExpectation{
				messagePrefix: "missing required scripts",
				availableParts: []string{
					"'script 37' at path 'path 36.require[0]'",
					"'script 40' at path 'path 38.require[0]'",
				},
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateScripts(tc.scripts)
			if tc.errorExpectation.messagePrefix == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if err == nil {
					return
				}

				assert.Contains(t, err.Error(), tc.errorExpectation.messagePrefix)

				for _, expectedMsgPart := range tc.errorExpectation.availableParts {
					assert.Contains(t, err.Error(), expectedMsgPart)
				}
			}
		})
	}
}

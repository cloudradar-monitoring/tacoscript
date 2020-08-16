package tasks

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cloudradar-monitoring/tacoscript/utils"

	"github.com/cloudradar-monitoring/tacoscript/conv"

	"github.com/stretchr/testify/assert"
)

func TestCmdRunTaskBuilder(t *testing.T) {
	testCases := []struct {
		typeName     string
		path         string
		ctx          []map[string]interface{}
		expectedTask *CmdRunTask
	}{
		{
			typeName: "someType",
			path:     "somePath",
			ctx: []map[string]interface{}{
				{
					NameField:  1,
					CwdField:   "somedir",
					UserField:  "someuser",
					ShellField: "someshell",
					EnvField: BuildExpectedEnvs(map[interface{}]interface{}{
						"one": "1",
						"two": "2",
					}),
					CreatesField: "somefile.txt",
					OnlyIf:       "one condition",
				},
			},
			expectedTask: &CmdRunTask{
				TypeName:   "someType",
				Path:       "somePath",
				Name:       "1",
				WorkingDir: "somedir",
				User:       "someuser",
				Shell:      "someshell",
				Envs: conv.KeyValues{
					{
						Key:   "one",
						Value: "1",
					},
					{
						Key:   "two",
						Value: "2",
					},
				},
				MissingFilesCondition: []string{"somefile.txt"},
				Errors:                &utils.Errors{},
				OnlyIf:                []string{"one condition"},
				FsManager:             &utils.FsManagerMock{},
			},
		},
		{
			typeName: "someTypeWithErrors",
			path:     "somePathWithErrors",
			ctx: []map[string]interface{}{
				{
					EnvField: 123,
				},
			},
			expectedTask: &CmdRunTask{
				TypeName: "someTypeWithErrors",
				Path:     "somePathWithErrors",
				Envs:     conv.KeyValues{},
				Errors: &utils.Errors{
					Errs: []error{
						fmt.Errorf("key value array expected at 'somePathWithErrors' but got '123'"),
					},
				},
				FsManager: &utils.FsManagerMock{},
			},
		},
		{
			typeName: "someTypeWithErrors2",
			path:     "somePathWithErrors2",
			ctx: []map[string]interface{}{
				{
					EnvField: []interface{}{
						"one",
					},
				},
			},
			expectedTask: &CmdRunTask{
				TypeName: "someTypeWithErrors2",
				Path:     "somePathWithErrors2",
				Envs:     conv.KeyValues{},
				Errors: &utils.Errors{
					Errs: []error{
						errors.New(`wrong key value element at 'somePathWithErrors2': '"one"'`),
					},
				},
				FsManager: &utils.FsManagerMock{},
			},
		},
		{
			typeName: "manyNamesType",
			path:     "manyNamesPath",
			ctx: []map[string]interface{}{
				{
					RequireField: "one require field",
					NamesField: []interface{}{
						"name one",
						"name two",
					},
				},
			},
			expectedTask: &CmdRunTask{
				TypeName: "manyNamesType",
				Path:     "manyNamesPath",
				Require: []string{
					"one require field",
				},
				Names: []string{
					"name one",
					"name two",
				},
				Errors:    &utils.Errors{},
				FsManager: &utils.FsManagerMock{},
			},
		},
		{
			typeName: "manyCreatesType",
			path:     "manyCreatesPath",
			ctx: []map[string]interface{}{
				{
					NameField: "many creates command",
					CreatesField: []interface{}{
						"create one",
						"create two",
						"create three",
					},
					RequireField: []interface{}{
						"req one",
						"req two",
						"req three",
					},
					OnlyIf: []interface{}{
						"OnlyIf one",
						"OnlyIf two",
						"OnlyIf three",
					},
				},
			},
			expectedTask: &CmdRunTask{
				TypeName: "manyCreatesType",
				Path:     "manyCreatesPath",
				Name:     "many creates command",
				Errors:   &utils.Errors{},
				MissingFilesCondition: []string{
					"create one",
					"create two",
					"create three",
				},
				Require: []string{
					"req one",
					"req two",
					"req three",
				},
				OnlyIf: []string{
					"OnlyIf one",
					"OnlyIf two",
					"OnlyIf three",
				},
				FsManager: &utils.FsManagerMock{},
			},
		},
		{
			typeName: "oneUnlessValue",
			path:     "oneUnlessValuePath",
			ctx: []map[string]interface{}{
				{
					NameField: "one unless value",
					Unless:    "unless one",
				},
			},
			expectedTask: &CmdRunTask{
				TypeName: "oneUnlessValue",
				Path:     "oneUnlessValuePath",
				Name:     "one unless value",
				Errors:   &utils.Errors{},
				Unless: []string{
					"unless one",
				},
				FsManager: &utils.FsManagerMock{},
			},
		},
		{
			typeName: "manyUnlessValue",
			path:     "manyUnlessValuePath",
			ctx: []map[string]interface{}{
				{
					NameField: "many unless value",
					Unless: []interface{}{
						"Unless one",
						"Unless two",
						"Unless three",
					},
				},
			},
			expectedTask: &CmdRunTask{
				TypeName: "manyUnlessValue",
				Path:     "manyUnlessValuePath",
				Name:     "many unless value",
				Errors:   &utils.Errors{},
				Unless: []string{
					"Unless one",
					"Unless two",
					"Unless three",
				},
				FsManager: &utils.FsManagerMock{},
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.typeName, func(t *testing.T) {
			cmdBuilder := CmdRunTaskBuilder{
				FsManager: &utils.FsManagerMock{},
			}
			actualTask, err := cmdBuilder.Build(
				tc.typeName,
				tc.path,
				tc.ctx,
			)

			assert.NoError(t, err)
			if err != nil {
				return
			}

			actualCmdRunTask, ok := actualTask.(*CmdRunTask)
			assert.True(t, ok)
			if !ok {
				return
			}

			assert.Equal(t, tc.expectedTask.User, actualCmdRunTask.User)
			AssertEnvValuesMatch(t, tc.expectedTask.Envs, actualCmdRunTask.Envs.ToEqualSignStrings())
			assert.Equal(t, tc.expectedTask.Path, actualCmdRunTask.Path)
			assert.Equal(t, tc.expectedTask.WorkingDir, actualCmdRunTask.WorkingDir)
			assert.Equal(t, tc.expectedTask.MissingFilesCondition, actualCmdRunTask.MissingFilesCondition)
			assert.Equal(t, tc.expectedTask.Name, actualCmdRunTask.Name)
			assert.Equal(t, tc.expectedTask.TypeName, actualCmdRunTask.TypeName)
			assert.Equal(t, tc.expectedTask.Shell, actualCmdRunTask.Shell)
			assert.Equal(t, tc.expectedTask.Names, actualCmdRunTask.Names)
			assert.Equal(t, tc.expectedTask.Require, actualCmdRunTask.Require)
			assert.Equal(t, tc.expectedTask.OnlyIf, actualCmdRunTask.OnlyIf)
			assert.Equal(t, tc.expectedTask.Unless, actualCmdRunTask.Unless)
			assert.EqualValues(t, tc.expectedTask.Errors, actualCmdRunTask.Errors)
		})
	}
}

func BuildExpectedEnvs(expectedEnvs map[interface{}]interface{}) []interface{} {
	envs := make([]interface{}, 0, len(expectedEnvs))
	for envKey, envValue := range expectedEnvs {
		envs = append(envs, map[interface{}]interface{}{
			envKey: envValue,
		})
	}

	return envs
}

func AssertEnvValuesMatch(t *testing.T, expectedEnvs conv.KeyValues, actualCmdEnvs []string) {
	expectedRawEnvs := expectedEnvs.ToEqualSignStrings()
	notFoundEnvs := make([]string, 0, len(expectedEnvs))
	for _, expectedRawEnv := range expectedRawEnvs {
		foundEnv := false
		for _, actualCmdEnv := range actualCmdEnvs {
			if expectedRawEnv == actualCmdEnv {
				foundEnv = true
				break
			}
		}

		if !foundEnv {
			notFoundEnvs = append(notFoundEnvs, expectedRawEnv)
		}
	}

	assert.Empty(
		t,
		notFoundEnvs,
		"was not able to find expected environment variables %s in cmd envs %s",
		strings.Join(notFoundEnvs, ", "),
		strings.Join(actualCmdEnvs, ", "),
	)
}

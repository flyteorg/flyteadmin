package common

import (
	"io/ioutil"
	"os"
	"testing"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestPopulateTemplateValues(t *testing.T) {
	const testEnvVarName = "TEST_FOO"
	tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix-")
	if err != nil {
		t.Fatalf("Cannot create temporary file: %v", err)
	}
	defer tmpFile.Close()
	_, err = tmpFile.WriteString("3")
	if err != nil {
		t.Fatalf("Cannot write to temporary file: %v", err)
	}

	data := map[string]runtimeInterfaces.DataSource{
		"directValue": {
			Value: "1",
		},
		"envValue": {
			ValueFrom: runtimeInterfaces.DataSourceValueFrom{
				EnvVar: testEnvVarName,
			},
		},
		"filePath": {
			ValueFrom: runtimeInterfaces.DataSourceValueFrom{
				FilePath: tmpFile.Name(),
			},
		},
	}
	origEnvVar := os.Getenv(testEnvVarName)
	defer os.Setenv(testEnvVarName, origEnvVar)
	os.Setenv(testEnvVarName, "2")

	templateValues, err := PopulateTemplateValues(data)
	assert.NoError(t, err)
	assert.EqualValues(t, map[string]string{
		"{{ directValue }}": "1",
		"{{ envValue }}":    "2",
		"{{ filePath }}":    "3",
	}, templateValues)
}

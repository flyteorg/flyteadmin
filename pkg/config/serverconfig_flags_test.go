// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

var dereferencableKindsServerConfig = map[reflect.Kind]struct{}{
	reflect.Array: {}, reflect.Chan: {}, reflect.Map: {}, reflect.Ptr: {}, reflect.Slice: {},
}

// Checks if t is a kind that can be dereferenced to get its underlying type.
func canGetElementServerConfig(t reflect.Kind) bool {
	_, exists := dereferencableKindsServerConfig[t]
	return exists
}

// This decoder hook tests types for json unmarshaling capability. If implemented, it uses json unmarshal to build the
// object. Otherwise, it'll just pass on the original data.
func jsonUnmarshalerHookServerConfig(_, to reflect.Type, data interface{}) (interface{}, error) {
	unmarshalerType := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	if to.Implements(unmarshalerType) || reflect.PtrTo(to).Implements(unmarshalerType) ||
		(canGetElementServerConfig(to.Kind()) && to.Elem().Implements(unmarshalerType)) {

		raw, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("Failed to marshal Data: %v. Error: %v. Skipping jsonUnmarshalHook", data, err)
			return data, nil
		}

		res := reflect.New(to).Interface()
		err = json.Unmarshal(raw, &res)
		if err != nil {
			fmt.Printf("Failed to umarshal Data: %v. Error: %v. Skipping jsonUnmarshalHook", data, err)
			return data, nil
		}

		return res, nil
	}

	return data, nil
}

func decode_ServerConfig(input, result interface{}) error {
	config := &mapstructure.DecoderConfig{
		TagName:          "json",
		WeaklyTypedInput: true,
		Result:           result,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			jsonUnmarshalerHookServerConfig,
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func join_ServerConfig(arr interface{}, sep string) string {
	listValue := reflect.ValueOf(arr)
	strs := make([]string, 0, listValue.Len())
	for i := 0; i < listValue.Len(); i++ {
		strs = append(strs, fmt.Sprintf("%v", listValue.Index(i)))
	}

	return strings.Join(strs, sep)
}

func testDecodeJson_ServerConfig(t *testing.T, val, result interface{}) {
	assert.NoError(t, decode_ServerConfig(val, result))
}

func testDecodeRaw_ServerConfig(t *testing.T, vStringSlice, result interface{}) {
	assert.NoError(t, decode_ServerConfig(vStringSlice, result))
}

func TestServerConfig_GetPFlagSet(t *testing.T) {
	val := ServerConfig{}
	cmdFlags := val.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())
}

func TestServerConfig_SetFlags(t *testing.T) {
	actual := ServerConfig{}
	cmdFlags := actual.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())

	t.Run("Test_httpPort", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("httpPort", testValue)
			if vInt, err := cmdFlags.GetInt("httpPort"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vInt), &actual.HTTPPort)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_grpcPort", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("grpcPort", testValue)
			if vInt, err := cmdFlags.GetInt("grpcPort"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vInt), &actual.GrpcPort)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_grpcServerReflection", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("grpcServerReflection", testValue)
			if vBool, err := cmdFlags.GetBool("grpcServerReflection"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vBool), &actual.GrpcServerReflection)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_kube-config", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("kube-config", testValue)
			if vString, err := cmdFlags.GetString("kube-config"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vString), &actual.KubeConfig)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_master", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("master", testValue)
			if vString, err := cmdFlags.GetString("master"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vString), &actual.Master)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_security.secure", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("security.secure", testValue)
			if vBool, err := cmdFlags.GetBool("security.secure"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vBool), &actual.Security.Secure)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_security.ssl.certificateFile", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("security.ssl.certificateFile", testValue)
			if vString, err := cmdFlags.GetString("security.ssl.certificateFile"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vString), &actual.Security.Ssl.CertificateFile)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_security.ssl.keyFile", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("security.ssl.keyFile", testValue)
			if vString, err := cmdFlags.GetString("security.ssl.keyFile"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vString), &actual.Security.Ssl.KeyFile)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_security.useAuth", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("security.useAuth", testValue)
			if vBool, err := cmdFlags.GetBool("security.useAuth"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vBool), &actual.Security.UseAuth)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_security.auditAccess", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("security.auditAccess", testValue)
			if vBool, err := cmdFlags.GetBool("security.auditAccess"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vBool), &actual.Security.AuditAccess)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_security.allowCors", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("security.allowCors", testValue)
			if vBool, err := cmdFlags.GetBool("security.allowCors"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vBool), &actual.Security.AllowCors)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_security.allowedOrigins", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := join_ServerConfig("1,1", ",")

			cmdFlags.Set("security.allowedOrigins", testValue)
			if vStringSlice, err := cmdFlags.GetStringSlice("security.allowedOrigins"); err == nil {
				testDecodeRaw_ServerConfig(t, join_ServerConfig(vStringSlice, ","), &actual.Security.AllowedOrigins)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_security.allowedHeaders", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := join_ServerConfig("1,1", ",")

			cmdFlags.Set("security.allowedHeaders", testValue)
			if vStringSlice, err := cmdFlags.GetStringSlice("security.allowedHeaders"); err == nil {
				testDecodeRaw_ServerConfig(t, join_ServerConfig(vStringSlice, ","), &actual.Security.AllowedHeaders)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_grpc.port", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("grpc.port", testValue)
			if vInt, err := cmdFlags.GetInt("grpc.port"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vInt), &actual.GrpcConfig.Port)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_grpc.serverReflection", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("grpc.serverReflection", testValue)
			if vBool, err := cmdFlags.GetBool("grpc.serverReflection"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vBool), &actual.GrpcConfig.ServerReflection)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_grpc.maxMessageSizeBytes", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("grpc.maxMessageSizeBytes", testValue)
			if vInt, err := cmdFlags.GetInt("grpc.maxMessageSizeBytes"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vInt), &actual.GrpcConfig.MaxMessageSizeBytes)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_thirdPartyConfig.flyteClient.clientId", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("thirdPartyConfig.flyteClient.clientId", testValue)
			if vString, err := cmdFlags.GetString("thirdPartyConfig.flyteClient.clientId"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vString), &actual.DeprecatedThirdPartyConfig.FlyteClientConfig.ClientID)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_thirdPartyConfig.flyteClient.redirectUri", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("thirdPartyConfig.flyteClient.redirectUri", testValue)
			if vString, err := cmdFlags.GetString("thirdPartyConfig.flyteClient.redirectUri"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vString), &actual.DeprecatedThirdPartyConfig.FlyteClientConfig.RedirectURI)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_thirdPartyConfig.flyteClient.scopes", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := join_ServerConfig("1,1", ",")

			cmdFlags.Set("thirdPartyConfig.flyteClient.scopes", testValue)
			if vStringSlice, err := cmdFlags.GetStringSlice("thirdPartyConfig.flyteClient.scopes"); err == nil {
				testDecodeRaw_ServerConfig(t, join_ServerConfig(vStringSlice, ","), &actual.DeprecatedThirdPartyConfig.FlyteClientConfig.Scopes)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_dataProxy.upload.maxSize", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := defaultServerConfig.DataProxy.Upload.MaxSize.String()

			cmdFlags.Set("dataProxy.upload.maxSize", testValue)
			if vString, err := cmdFlags.GetString("dataProxy.upload.maxSize"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vString), &actual.DataProxy.Upload.MaxSize)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_dataProxy.upload.maxExpiresIn", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := defaultServerConfig.DataProxy.Upload.MaxExpiresIn.String()

			cmdFlags.Set("dataProxy.upload.maxExpiresIn", testValue)
			if vString, err := cmdFlags.GetString("dataProxy.upload.maxExpiresIn"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vString), &actual.DataProxy.Upload.MaxExpiresIn)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_dataProxy.upload.defaultFileNameLength", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("dataProxy.upload.defaultFileNameLength", testValue)
			if vInt, err := cmdFlags.GetInt("dataProxy.upload.defaultFileNameLength"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vInt), &actual.DataProxy.Upload.DefaultFileNameLength)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_dataProxy.upload.storagePrefix", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("dataProxy.upload.storagePrefix", testValue)
			if vString, err := cmdFlags.GetString("dataProxy.upload.storagePrefix"); err == nil {
				testDecodeJson_ServerConfig(t, fmt.Sprintf("%v", vString), &actual.DataProxy.Upload.StoragePrefix)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
}

package autoflags

import (
	"flag"
	"net/url"
	"testing"

	"github.com/buildbuddy-io/buildbuddy/server/util/flagutil/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flagtypes "github.com/buildbuddy-io/buildbuddy/server/util/flagutil/types"
	flagtags "github.com/buildbuddy-io/buildbuddy/server/util/flagutil/types/autoflags/tags"
	flagyaml "github.com/buildbuddy-io/buildbuddy/server/util/flagutil/yaml"
)

type unsupportedFlagValue struct{}

func (f *unsupportedFlagValue) Set(string) error { return nil }
func (f *unsupportedFlagValue) String() string   { return "" }

type testStruct struct {
	Field  int    `json:"field"`
	Meadow string `json:"meadow"`
}

func replaceFlagsForTesting(t *testing.T) *flag.FlagSet {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	common.DefaultFlagSet = flags

	t.Cleanup(func() {
		common.DefaultFlagSet = flag.CommandLine
	})

	return flags
}

func replaceIgnoreSetForTesting(t *testing.T) map[string]struct{} {
	oldIgnoreSet := flagyaml.IgnoreSet
	flagyaml.IgnoreSet = make(map[string]struct{})

	t.Cleanup(func() {
		flagyaml.IgnoreSet = oldIgnoreSet
	})

	return flagyaml.IgnoreSet
}

func TestNew(t *testing.T) {
	flags := replaceFlagsForTesting(t) //nolint:SA4006
	flagBool := New(flags, "bool", false, "")
	err := common.SetValueForFlagName(flags, "bool", true, map[string]struct{}{}, true)
	require.NoError(t, err)
	assert.Equal(t, true, *flagBool)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagBool = New(flags, "bool", false, "")
	err = common.SetValueForFlagName(flags, "bool", true, map[string]struct{}{"bool": {}}, true)
	require.NoError(t, err)
	assert.Equal(t, false, *flagBool)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagInt := New(flags, "int", 2, "")
	err = common.SetValueForFlagName(flags, "int", 1, map[string]struct{}{}, true)
	require.NoError(t, err)
	assert.Equal(t, 1, *flagInt)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagInt = New(flags, "int", 2, "")
	err = common.SetValueForFlagName(flags, "int", 1, map[string]struct{}{"int": {}}, true)
	require.NoError(t, err)
	assert.Equal(t, 2, *flagInt)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagInt64 := New(flags, "int64", int64(2), "")
	err = common.SetValueForFlagName(flags, "int64", 1, map[string]struct{}{}, true)
	require.NoError(t, err)
	assert.Equal(t, int64(1), *flagInt64)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagInt64 = New(flags, "int64", int64(2), "")
	err = common.SetValueForFlagName(flags, "int64", 1, map[string]struct{}{"int64": {}}, true)
	require.NoError(t, err)
	assert.Equal(t, int64(2), *flagInt64)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagUint := New(flags, "uint", uint(2), "")
	err = common.SetValueForFlagName(flags, "uint", 1, map[string]struct{}{}, true)
	require.NoError(t, err)
	assert.Equal(t, uint(1), *flagUint)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagUint = New(flags, "uint", uint(2), "")
	err = common.SetValueForFlagName(flags, "uint", 1, map[string]struct{}{"uint": {}}, true)
	require.NoError(t, err)
	assert.Equal(t, uint(2), *flagUint)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagUint64 := New(flags, "uint64", uint64(2), "")
	err = common.SetValueForFlagName(flags, "uint64", 1, map[string]struct{}{}, true)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), *flagUint64)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagUint64 = New(flags, "uint64", uint64(2), "")
	err = common.SetValueForFlagName(flags, "uint64", 1, map[string]struct{}{"uint64": {}}, true)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), *flagUint64)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagFloat64 := New(flags, "float64", float64(2), "")
	err = common.SetValueForFlagName(flags, "float64", 1, map[string]struct{}{}, true)
	require.NoError(t, err)
	assert.Equal(t, float64(1), *flagFloat64)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagFloat64 = New(flags, "float64", float64(2), "")
	err = common.SetValueForFlagName(flags, "float64", 1, map[string]struct{}{"float64": {}}, true)
	require.NoError(t, err)
	assert.Equal(t, float64(2), *flagFloat64)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagString := New(flags, "string", "2", "")
	err = common.SetValueForFlagName(flags, "string", "1", map[string]struct{}{}, true)
	require.NoError(t, err)
	assert.Equal(t, "1", *flagString)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagString = New(flags, "string", "2", "")
	err = common.SetValueForFlagName(flags, "string", "1", map[string]struct{}{"string": {}}, true)
	require.NoError(t, err)
	assert.Equal(t, "2", *flagString)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	defaultURL := url.URL{Scheme: "https", Host: "www.example.com"}
	flagURL := New(flags, "url", defaultURL, "")
	u, err := url.Parse("https://www.example.com:8080")
	require.NoError(t, err)
	err = common.SetValueForFlagName(flags, "url", *u, map[string]struct{}{}, true)
	require.NoError(t, err)
	assert.Equal(t, url.URL{Scheme: "https", Host: "www.example.com"}, defaultURL)
	assert.Equal(t, url.URL{Scheme: "https", Host: "www.example.com:8080"}, *flagURL)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	flagURL = New(flags, "url", url.URL{Scheme: "https", Host: "www.example.com"}, "")
	u, err = url.Parse("https://www.example.com:8080")
	require.NoError(t, err)
	err = common.SetValueForFlagName(flags, "url", *u, map[string]struct{}{"url": {}}, true)
	require.NoError(t, err)
	assert.Equal(t, url.URL{Scheme: "https", Host: "www.example.com"}, *flagURL)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	defaultStringSlice := []string{"1", "2"}
	stringSlice := New(flags, "string_slice", defaultStringSlice, "")
	err = common.SetValueForFlagName(flags, "string_slice", []string{"3", "4", "5", "6", "7", "8", "9", "0", "1", "2"}, map[string]struct{}{}, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, defaultStringSlice)
	assert.Equal(t, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "1", "2"}, *stringSlice)

	flags = replaceFlagsForTesting(t)
	stringSlice = New(flags, "string_slice", defaultStringSlice, "")
	err = common.SetValueForFlagName(flags, "string_slice", []string{"3"}, map[string]struct{}{"string_slice": {}}, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, defaultStringSlice)
	assert.Equal(t, []string{"1", "2", "3"}, *(*[]string)(flags.Lookup("string_slice").Value.(*flagtypes.StringSliceFlag)))
	assert.Equal(t, []string{"1", "2", "3"}, *stringSlice)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	stringSlice = New(flags, "string_slice", defaultStringSlice, "")
	err = common.SetValueForFlagName(flags, "string_slice", []string{"3"}, map[string]struct{}{}, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, defaultStringSlice)
	assert.Equal(t, []string{"3"}, *stringSlice)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	stringSlice = New(flags, "string_slice", defaultStringSlice, "")
	err = common.SetValueForFlagName(flags, "string_slice", []string{"3"}, map[string]struct{}{"string_slice": {}}, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, defaultStringSlice)
	assert.Equal(t, []string{"1", "2"}, *stringSlice)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	defaultStructSlice := []testStruct{{Field: 1}, {Field: 2}}
	structSlice := New(flags, "struct_slice", defaultStructSlice, "")
	err = common.SetValueForFlagName(flags, "struct_slice", []testStruct{{Field: 3}}, map[string]struct{}{}, true)
	require.NoError(t, err)
	assert.Equal(t, []testStruct{{Field: 1}, {Field: 2}, {Field: 3}}, *structSlice)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	structSlice = New(flags, "struct_slice", defaultStructSlice, "")
	err = common.SetValueForFlagName(flags, "struct_slice", []testStruct{{Field: 3}}, map[string]struct{}{"struct_slice": {}}, true)
	require.NoError(t, err)
	assert.Equal(t, []testStruct{{Field: 1}, {Field: 2}}, defaultStructSlice)
	assert.Equal(t, []testStruct{{Field: 1}, {Field: 2}, {Field: 3}}, *structSlice)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	structSlice = New(flags, "struct_slice", defaultStructSlice, "")
	err = common.SetValueForFlagName(flags, "struct_slice", []testStruct{{Field: 3}}, map[string]struct{}{}, false)
	require.NoError(t, err)
	assert.Equal(t, []testStruct{{Field: 3}}, *structSlice)

	flags = replaceFlagsForTesting(t) //nolint:SA4006
	structSlice = New(flags, "struct_slice", defaultStructSlice, "")
	err = common.SetValueForFlagName(flags, "struct_slice", []testStruct{{Field: 3}}, map[string]struct{}{"struct_slice": {}}, false)
	require.NoError(t, err)
	assert.Equal(t, []testStruct{{Field: 1}, {Field: 2}}, *structSlice)
}

func TestYAMLIgnoreTag(t *testing.T) {
	flags := replaceFlagsForTesting(t)
	_ = replaceIgnoreSetForTesting(t)
	flagBool := New(flags, "bool", false, "", flagtags.YAMLIgnoreTag)

	yamlData := `
	bool: true
`
	flagyaml.PopulateFlagsFromData(yamlData)
	assert.Equal(t, false, *flagBool)
}

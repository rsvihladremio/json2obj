/*
Copyright 2022 Ryan SVIHLA

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

// GitSha is added from the build and release scripts
var GitSha = "unknown"

// Version is pulled from the branch name and set in the build and release scripts
var Version = "unknownVersion"

var platform = runtime.GOOS
var arch = runtime.GOARCH

var Lang string
var Output string

func PrintHeader(version, platform, arch, gitSha string) string {
	return fmt.Sprintf("json2obj %v-%v-%v-%v\n", version, gitSha, platform, arch)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "json2obj [flags] file.json",
	Short: "converts json to code",
	Long:  `converts a json text file to a data object in the language of your choice`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Optionally run one of the validators provided by cobra
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		PrintHeader(Version, platform, arch, GitSha)

		jsonBytes, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Printf("unable to open file due to error: '%v'\n", err)
			os.Exit(1)
		}

		var result map[string]interface{}
		err = json.Unmarshal(jsonBytes, &result)
		if err != nil {
			fmt.Printf("unable to marshall json due to error: '%v'\n", err)
			os.Exit(1)
		}
		builder := strings.Builder{}
		name := "JsonObject"
		if Output != "" {
			tokens := strings.Split(Output, ".")
			name = tokens[0]
		}
		if Lang == "java" {
			_, err = builder.WriteString("package com.example;\nimport java.util.List;\nimport java.util.Map;\n\npublic class ")
			if err != nil {
				fmt.Printf("unable to write class declaration due to error: '%v'\n", err)
				os.Exit(1)
			}
			_, err = builder.WriteString(name)
			if err != nil {
				fmt.Printf("unable to write class name due to error: '%v'\n", err)
				os.Exit(1)
			}
			_, err = builder.WriteString(" {\n\n")
			if err != nil {
				fmt.Printf("unable to write closing class declaration due to error: '%v'\n", err)
				os.Exit(1)
			}
			for k, v := range result {
				str, err := writeJava(k, v)
				if err != nil {
					fmt.Printf("error generating java code for %v with error: '%v'\n", v, err)
					os.Exit(1)
				}
				_, err = builder.WriteString(str)
				if err != nil {
					fmt.Printf("unable to write key %v due to error: '%v'\n", k, err)
					os.Exit(1)
				}
			}
			_, err = builder.WriteString("}\n")
			if err != nil {
				fmt.Printf("unable to write string due to error: '%v'\n", err)
				os.Exit(1)
			}
		}
		if Output == "" {
			fmt.Println(builder.String())
		} else {
			os.WriteFile(Output, []byte(builder.String()), 0755)
		}

	},
}

func writeJava(key string, v interface{}) (string, error) {
	var typeName string = ""
	vType := reflect.ValueOf(v)

	switch vType.Kind() {
	case reflect.Float32:
		typeName = "float"
	case reflect.Float64:
		typeName = "double"
	case reflect.Int:
		typeName = "int"
	case reflect.String:
		typeName = "String"
	case reflect.Bool:
		typeName = "boolean"
	case reflect.Slice:
		typeName = "List"
	case reflect.Map:
		typeName = "Map<String, Object>"
	default:
		return "", fmt.Errorf("unable to handle type %t for %v", v, key)
	}
	field := fmt.Sprintf("\tprivate %v %v;\n\n", typeName, key)
	setter := fmt.Sprintf("\tpublic void set%v(%v %v){\n\t\tthis.%v = %v;\n\t}\n\n", capitalize(key), typeName, key, key, key)
	getter := fmt.Sprintf("\tpublic %v get%v(){\n\t\treturn this.%v;\n\t}\n\n", typeName, capitalize(key), key)
	return field + setter + getter, nil
}

func capitalize(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&Lang, "lang", "l", "go", "output language")
	rootCmd.Flags().StringVarP(&Output, "output", "o", "", "output file")
}

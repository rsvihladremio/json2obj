// Copyright 2022 Ryan SVIHLA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmd is where the logic of the command line flags is
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// Lang is the language to generate output on
var Lang string

// Output is where the file will be written
var Output string

// PrintHeader provides the program version
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
		return cobra.MinimumNArgs(1)(cmd, args)
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
		name := "JsonObject"
		if Output != "" {
			rawName := capitalize(filepath.Base(Output))
			tokens := strings.Split(rawName, ".")
			name = tokens[0]
		}
		var classText string
		if Lang == "java" {

			header := "package com.example;\nimport java.util.List;\nimport java.util.Map;\n\n"
			classTextRaw, err := writeJavaClass(name, false, result)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			classText = fmt.Sprintf("%v\n%v", header, classTextRaw)
		} else if Lang == "go" {
			header := "package test;\n\nimport (\n\"fmt\"\n)\n"
			classTextRaw, err := writeGoStruct(name, result)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			classText = fmt.Sprintf("%v\n%v", header, classTextRaw)
		} else {
			fmt.Printf("unknown programming language %v", Lang)
			os.Exit(1)
		}
		if Output == "" {
			fmt.Println(classText)
		} else {
			os.WriteFile(Output, []byte(classText), 0755)
		}

	},
}

func writeGoStruct(name string, result map[string]interface{}) (string, error) {
	builder := strings.Builder{}
	_, err := builder.WriteString(fmt.Sprintf("type %v struct{\n", name))
	if err != nil {
		return "", fmt.Errorf("unable to write struct name due to error: '%v", err)
	}
	var nestedClassesArray []string
	for k, v := range result {
		str, nestedClasses, err := writeGo(k, v)
		if err != nil {
			return "", fmt.Errorf("error generating go code for %v with error: '%v'", v, err)
		}
		_, err = builder.WriteString(str)
		if err != nil {
			return "", fmt.Errorf("unable to write key %v due to error: '%v'", k, err)
		}
		if nestedClasses != "" {
			nestedClassesArray = append(nestedClassesArray, nestedClasses)
		}
	}
	_, err = builder.WriteString("}\n")
	if err != nil {
		return "", fmt.Errorf("unable to write string due to error: '%v'", err)
	}

	for _, nested := range nestedClassesArray {
		_, err = builder.WriteString(nested)
		if err != nil {
			return "", fmt.Errorf("unable to write nested struct %v due to error: '%v'", nested, err)
		}
	}
	return builder.String(), nil
}

func writeGo(key string, v interface{}) (fieldText string, nestedClasses string, err error) {
	var nestedClassesArray []string
	var typeName string
	vType := reflect.ValueOf(v)
	switch vType.Kind() {
	case reflect.Float64:
		//go always unmarshalls numbers to this so we are going to detect things
		if strings.Contains(fmt.Sprintf("%v", v), ".") {
			typeName = "float64"
		} else {
			typeName = "int64"
		}
	case reflect.String:
		typeName = "string"
	case reflect.Bool:
		typeName = "bool"
	case reflect.Slice:
		sliceValue := v.([]interface{})
		if len(sliceValue) > 0 {
			firstElement := sliceValue[0]
			firstElementType := reflect.ValueOf(firstElement)
			switch firstElementType.Kind() {
			case reflect.Float32:
				typeName = "[]float32"
			case reflect.Float64:
				typeName = "[]float64"
			case reflect.Int:
				typeName = "[]int"
			case reflect.String:
				typeName = "[]string"
			case reflect.Bool:
				typeName = "[]bool"
			case reflect.Slice:
				typeName = "[]interface{}"
			case reflect.Map:
				// make a nested class
				mapValue := firstElement.(map[string]interface{})
				nestedValueName := fmt.Sprintf("%v", capitalize(key))
				newNestedClassStr, err := writeGoStruct(nestedValueName, mapValue)
				if err != nil {
					return "", "", fmt.Errorf("unable to handle nested type %T for %v", mapValue, key)
				}
				nestedClassesArray = append(nestedClassesArray, newNestedClassStr)
				typeName = fmt.Sprintf("[]%v", nestedValueName)
			}
		} else {
			// no types so we can't guess just assume this
			typeName = "[]interface{}"
		}
	case reflect.Map:
		mapValue := v.(map[string]interface{})
		nestedValueName := fmt.Sprintf("%v", capitalize(key))
		newNestedClassStr, err := writeGoStruct(nestedValueName, mapValue)
		if err != nil {
			return "", "", fmt.Errorf("unable to handle nested type %T for %v", mapValue, key)
		}
		nestedClassesArray = append(nestedClassesArray, newNestedClassStr)
		typeName = nestedValueName
	default:
		return "", "", fmt.Errorf("unable to handle type %t for %v", v, key)
	}
	field := fmt.Sprintf("\t%v %v `json:\"%v\"`\n", capitalize(key), typeName, key)
	return field, strings.Join(nestedClassesArray, "\n\n"), nil
}

func writeJavaClass(name string, isStatic bool, result map[string]interface{}) (string, error) {
	staticString := ""
	if isStatic {
		staticString = "static "
	}
	builder := strings.Builder{}
	_, err := builder.WriteString(fmt.Sprintf("public %vclass %v {\n\n", staticString, name))
	if err != nil {
		return "", fmt.Errorf("unable to write class name due to error: '%v", err)
	}
	var nestedClassesArray []string
	for k, v := range result {
		str, nestedClasses, err := writeJava(k, v)
		if err != nil {
			return "", fmt.Errorf("error generating java code for %v with error: '%v'", v, err)
		}
		_, err = builder.WriteString(str)
		if err != nil {
			return "", fmt.Errorf("unable to write key %v due to error: '%v'", k, err)
		}
		if nestedClasses != "" {
			nestedClassesArray = append(nestedClassesArray, nestedClasses)
		}
	}

	for _, nested := range nestedClassesArray {
		_, err = builder.WriteString(nested)
		if err != nil {
			return "", fmt.Errorf("unable to write nested classes %v due to error: '%v'", nested, err)
		}
	}
	_, err = builder.WriteString("}\n")
	if err != nil {
		return "", fmt.Errorf("unable to write string due to error: '%v'", err)
	}
	return builder.String(), nil
}

func writeJava(key string, v interface{}) (fieldText string, nestedClasses string, err error) {
	var nestedClassesArray []string
	var typeName string
	vType := reflect.ValueOf(v)
	switch vType.Kind() {
	case reflect.Float64:
		//go always unmarshalls numbers to this so we are going to detect things
		if strings.Contains(fmt.Sprintf("%v", v), ".") {
			typeName = "double"
		} else {
			typeName = "long"
		}
	case reflect.String:
		typeName = "String"
	case reflect.Bool:
		typeName = "boolean"
	case reflect.Slice:
		sliceValue := v.([]interface{})
		if len(sliceValue) > 0 {
			firstElement := sliceValue[0]
			firstElementType := reflect.ValueOf(firstElement)
			switch firstElementType.Kind() {
			case reflect.Float32:
				typeName = "List<Float>"
			case reflect.Float64:
				typeName = "List<Double>"
			case reflect.Int:
				typeName = "List<Integer>"
			case reflect.String:
				typeName = "List<String>"
			case reflect.Bool:
				typeName = "List<Boolean>"
			case reflect.Slice:
				//too hard to do for now
				typeName = "List"
			case reflect.Map:
				// make a nested class
				mapValue := firstElement.(map[string]interface{})
				nestedValueName := fmt.Sprintf("%vNested1", capitalize(key))
				newNestedClassStr, err := writeJavaClass(nestedValueName, true, mapValue)
				if err != nil {
					return "", "", fmt.Errorf("unable to handle nested type %T for %v", mapValue, key)
				}
				nestedClassesArray = append(nestedClassesArray, newNestedClassStr)
				typeName = fmt.Sprintf("List<%v>", nestedValueName)
			}
		} else {
			// no types so we can't guess just assume this
			typeName = "List"
		}
	case reflect.Map:
		mapValue := v.(map[string]interface{})
		nestedValueName := fmt.Sprintf("%vNested1", capitalize(key))
		newNestedClassStr, err := writeJavaClass(nestedValueName, true, mapValue)
		if err != nil {
			return "", "", fmt.Errorf("unable to handle nested type %T for %v", mapValue, key)
		}
		nestedClassesArray = append(nestedClassesArray, newNestedClassStr)
		typeName = nestedValueName
	default:
		return "", "", fmt.Errorf("unable to handle type %t for %v", v, key)
	}
	field := fmt.Sprintf("\tprivate %v %v;\n\n", typeName, key)
	setter := fmt.Sprintf("\tpublic void set%v(%v %v){\n\t\tthis.%v = %v;\n\t}\n\n", capitalize(key), typeName, key, key, key)
	getter := fmt.Sprintf("\tpublic %v get%v(){\n\t\treturn this.%v;\n\t}\n\n", typeName, capitalize(key), key)
	return field + setter + getter, strings.Join(nestedClassesArray, "\n"), nil
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

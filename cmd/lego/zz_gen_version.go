// Code generated by 'internal/releaser'; DO NOT EDIT.

package main

const defaultVersion = "v4.23.1+dev-release"

var version = ""

func getVersion() string {
	if version == "" {
		return defaultVersion
	}

	return version
}

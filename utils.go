package main

import (
	"os"

	"github.com/v-braun/go-must"
)

func pathExists(filePath string) bool {
	var err error
	if _, err = os.Stat(filePath); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}

	must.NoError(err, "unexpected error in os.stat")
	return false
}

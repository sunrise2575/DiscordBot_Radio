package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/seehuhn/mt19937"
)

func readFileAsString(path string) string {
	out, e := ioutil.ReadFile(path)
	if e != nil {
		panic(e)
	}
	return string(out)
}

func splitFilepath(path string) (string, string, string) {
	absPath, _ := filepath.Abs(path)
	folder, base := filepath.Split(absPath)
	ext := filepath.Ext(absPath)
	name := strings.TrimSuffix(base, ext)

	return folder, name, ext
}

func findFilesInFolderRecursive(folderPath string) []string {
	result := []string{}

	folderPathAbs, e := filepath.Abs(folderPath)
	if e != nil {
		log.Println(e)
		return nil
	}

	e = filepath.Walk(folderPathAbs, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return nil
		}
		_, _, ext := splitFilepath(path)

		// filter file extension
		if !(ext == ".ogg" || ext == ".mp3" || ext == ".wav" || ext == ".flac" || ext == ".aac" || ext == ".mka") {
			return nil
		}

		result = append(result, path)
		return nil
	})

	if e != nil {
		log.Println(e)
		return nil
	}

	return result
}

func getRandomInt(max int) int {
	// select file
	rng := rand.New(mt19937.New())
	rng.Seed(time.Now().UnixNano())

	return rng.Intn(max)
}

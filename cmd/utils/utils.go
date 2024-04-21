package utils

import (
	"io"
	"net/http"
	"os"
	"regexp"
)

func MatchAllStringByRegex(regex, str string) ([]*string, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	result := r.FindAllSubmatch([]byte(str), -1)

	if len(result) < 1 {
		return nil, nil
	}

	var data []*string

	for _, v := range result {
		d := string(v[1])
		data = append(data, &d)
	}

	return data, nil
}

func DownloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func DoesDirectoryExistAndIsNotEmpty(name string) bool {
	if _, err := os.Stat(name); err == nil {
		dirEntries, err := os.ReadDir(name)
		if err != nil {
			return false
		}
		if len(dirEntries) > 0 {
			return true
		}
	}
	return false
}

func CopyFile(sourcePath, destinationPath string) error {
	// Open the source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	// Optionally, you can also flush the destination file to ensure all data is written
	err = destinationFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

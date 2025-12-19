package cache

import (
	"io"
)

func Valid(gopherfile string, directory string, goBin string) (bool, error) {
	// TODO: Implement caching here
	// If directory does not exist -> false
	// If gopherfile does not exist -> false
	// If goBin does not exist -> err
	return false, nil
}

type HashFiles struct {
	GopherFile string `json:"gopherfile"`
	TargetFile string `json:"targets.go"`
	GoMod      string `json:"go.mod"`
	GoSumm     string `json:"go.sum"`
}

type HashFileContents struct {
	GopherVersion string    `json:"gopher_version"`
	Hashes        HashFiles `json:"hashes"`
}

func HashFrom(reader io.Reader) (string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return Hash(string(content)), nil
}

func Hash(content string) string {
	// TODO: Implement a better hash function
	return content
}

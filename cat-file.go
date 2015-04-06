package gitgo

import (
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type KeyType string

type SHA [sha1.Size]byte

// NewSha is a convenience function that
// converts a hex string representation of a SHA
// to a SHA type.
// Its behavior is undefined when given an invalid
// hex string
func NewSha(str string) (sha [sha1.Size]byte) {

	bts, _ := hex.DecodeString(str)

	for i := 0; i < len(bts); i++ {
		sha[i] = bts[i]
	}
	return
}

const (
	TreeKey      KeyType = "tree"
	ParentKey            = "parent"
	AuthorKey            = "author"
	CommitterKey         = "committer"
)

type GitObject struct {
	Type string

	Tree      string
	Parents   []string
	Author    string
	Committer string
	Message   string
	Size      string
}

func NewObject(inputSha string) (obj GitObject, err error) {
	str, err := CatFile(inputSha)
	if err != nil {
		return
	}
	return parseObj(str)
}

func CatFile(inputSha string) (result string, err error) {

	filename := path.Join(".git", "objects", inputSha[:2], inputSha[2:])

	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	r, err := zlib.NewReader(f)
	if err != nil {
		return
	}

	bts, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}

	return string(bts), nil
}

func parseObj(obj string) (result GitObject, err error) {

	parts := strings.Split(obj, "\x00")
	parts = strings.Fields(parts[0])
	result.Type = parts[0]
	result.Size = parts[1]
	nullIndex := strings.Index(obj, "\x00")

	obj = obj[nullIndex+1:]
	lines := strings.Split(obj, "\n")

	for i, line := range lines {
		// The next line is the commit message
		if len(strings.Fields(line)) == 0 {
			result.Message = strings.Join(lines[i+1:], "\n")
			break
		}
		parts := strings.Fields(line)
		key := parts[0]
		switch KeyType(key) {
		case TreeKey:
			result.Tree = parts[1]
		case ParentKey:
			result.Parents = append(result.Parents, parts[1])
		case AuthorKey:
			result.Author = strings.Join(parts[1:], " ")
		case CommitterKey:
			result.Committer = strings.Join(parts[1:], " ")
		default:
			err = fmt.Errorf("Encounterd unknown field in commit: %s", key)
			return
		}
	}
	return
}
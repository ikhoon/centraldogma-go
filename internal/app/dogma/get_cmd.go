// Copyright 2018 LINE Corporation
//
// LINE Corporation licenses this file to you under the Apache License,
// version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at:
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"

	"github.com/urfave/cli"
	"go.linecorp.com/centraldogma"
)

// A getFileCommand fetches the content of the file in the specified path matched by the
// JSON path expressions with the specified revision.
type getFileCommand struct {
	out           io.Writer
	repo          repositoryRequestInfo
	localFilePath string
	jsonPaths     []string
}

func (gf *getFileCommand) execute(c *cli.Context) error {
	repo := gf.repo
	entry, err := getRemoteFileEntry(
		c, repo.remoteURL, repo.projName, repo.repoName, repo.path, repo.revision, gf.jsonPaths)
	if err != nil {
		return err
	}

	filePath := creatableFilePath(gf.localFilePath, 1)
	fd, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	if entry.Type == centraldogma.JSON {
		b := safeMarshalIndent(entry.Content)
		if _, err = fd.Write(b); err != nil {
			return err
		}
	} else if entry.Type == centraldogma.Text {
		_, err = fd.Write(entry.Content)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(gf.out, "Downloaded: %s\n", path.Base(filePath))
	return nil
}

// A catFileCommand shows the content of the file in the specified path matched by the
// JSON path expressions with the specified revision.
type catFileCommand struct {
	out       io.Writer
	repo      repositoryRequestInfo
	jsonPaths []string
}

func (cf *catFileCommand) execute(c *cli.Context) error {
	repo := cf.repo
	entry, err := getRemoteFileEntry(
		c, repo.remoteURL, repo.projName, repo.repoName, repo.path, repo.revision, cf.jsonPaths)
	if err != nil {
		return err
	}

	if entry.Type == centraldogma.JSON {
		b := safeMarshalIndent(entry.Content)
		fmt.Fprintf(cf.out, "%s\n", string(b))
	} else if entry.Type == centraldogma.Text { //
		fmt.Fprintf(cf.out, "%s\n", string(entry.Content))
	}

	return nil
}

func creatableFilePath(filePath string, inc int) string {
	regex, _ := regexp.Compile(`\.[0-9]*$`)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		if inc == 1 {
			filePath += "."
		}
		startIndex := regex.FindStringIndex(filePath)
		filePath = filePath[0:startIndex[0]+1] + strconv.Itoa(inc)
		inc++
		return creatableFilePath(filePath, inc)
	}
	return filePath
}

// newGetCommand creates the getCommand. If the localFilePath is not specified, the file name of the path
// will be set by default.
func newGetCommand(c *cli.Context, out io.Writer) (Command, error) {
	repo, err := newRepositoryRequestInfo(c)
	if err != nil {
		return nil, err
	}

	localFilePath := path.Base(repo.path)
	if len(c.Args()) == 2 && len(c.Args().Get(1)) != 0 {
		localFilePath = c.Args().Get(1)
	}

	return &getFileCommand{out: out, repo: repo, localFilePath: localFilePath, jsonPaths: c.StringSlice("jsonpath")}, nil
}

// newCatCommand creates the catCommand.
func newCatCommand(c *cli.Context, out io.Writer) (Command, error) {
	repo, err := newRepositoryRequestInfo(c)
	if err != nil {
		return nil, err
	}
	return &catFileCommand{out: out, repo: repo, jsonPaths: c.StringSlice("jsonpath")}, nil
}

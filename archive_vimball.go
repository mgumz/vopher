package main

// http://www.drchip.org/astronaut/vim/doc/pi_vimball.txt.html

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type vimball_efunc func(name string, lines int, s *bufio.Scanner) error

type VimballArchive struct{}

func (vimball *VimballArchive) Extract(folder string, r io.Reader, skip_dir int) error {

	f := func(n string, l int, s *bufio.Scanner) error {
		return vimball.extractFile(n, l, s)
	}
	_, err := vimball.handle(folder, r, f)
	return err
}

func (vimball *VimballArchive) Entries(r io.Reader, skip_dir int) ([]string, error) {
	f := func(n string, l int, s *bufio.Scanner) error {
		return vimball.skipFile(n, l, s)
	}
	return vimball.handle("", r, f)
}

// extracts the contents a vimball formatted 'r' into 'dir'
// the format works like this:
//
// preamble
//     " Vimball Archive by Charles E. Campbell, Jr. Ph.D.
//     UseVimball
//     finish
// file-contents
//     folder/name_of_file
//     number_of_lines
//     ...
//     ...
//     folder2/other_file
//     number_of_lines2
//     ...
//     ...
func (*VimballArchive) handle(folder string, r io.Reader, extract vimball_efunc) ([]string, error) {

	var (
		contents = make([]string, 0)
		scanner  = bufio.NewScanner(r)

		pre = struct{ use_vimball, finish bool }{} // 'preamble'
	)

	scanner.Split(bufio.ScanLines)

	// scan for lines 'UseVimball', followed by 'finish'
	for scanner.Scan() && !pre.use_vimball && !pre.finish {
		line := scanner.Text()
		if !pre.use_vimball && line == "UseVimball" {
			pre.use_vimball = true
		} else if pre.use_vimball && !pre.finish && line == "finish" {
			pre.finish = true
		}
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	if !pre.use_vimball && !pre.finish {
		return nil, fmt.Errorf("error vimball: strange preamble")
	}

	// now scan the file-entries
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.SplitN(line, "\t", 2)

		name := strings.TrimSpace(fields[0])
		name = filepath.Clean(name)

		if !scanner.Scan() {
			return nil, fmt.Errorf("error vimball: while scanning line-number for %q: %v", name, scanner.Err())
		}

		nlines, err := strconv.Atoi(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("error vimball: while parsing line-number for %q: %v", name, err)
		}
		if nlines < 0 {
			return nil, fmt.Errorf("error vimball: got negative line-number for %q: %v", name, nlines)
		}

		if err = extract(filepath.Join(folder, name), nlines, scanner); err != nil {
			return nil, err
		}
		contents = append(contents, name)
	}

	return contents, scanner.Err()
}

//
func (*VimballArchive) extractFile(name string, lines int, scanner *bufio.Scanner) error {

	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}

	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	for lines > 0 && scanner.Scan() {
		file.Write(scanner.Bytes())
		lines--
	}

	return scanner.Err()
}

func (*VimballArchive) skipFile(name string, lines int, scanner *bufio.Scanner) error {
	for lines > 0 && scanner.Scan() {
		lines--
	}
	return scanner.Err()
}

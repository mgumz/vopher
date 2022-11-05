package archive

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mgumz/vopher/pkg/vopher"
)

// VimballArchive handles vimball archives, see
// http://www.drchip.org/astronaut/vim/doc/pi_vimball.txt.html
type VimballArchive struct{}

func init() {

	suffixes := []string{".vba", ".vba.gz", ".vmb", ".vmb.gz"}
	supportedArchives = append(supportedArchives, suffixes...)

	archiveGuesser = append(archiveGuesser, func(n string) vopher.Archive {
		for _, s := range suffixes {
			if strings.HasSuffix(n, s) {
				if strings.HasSuffix(n, ".gz") {
					return &GzArchive{&VimballArchive{}}
				}
				return &VimballArchive{}
			}
		}
		return nil
	})
}

type vimballExtractFunc func(name string, lines int, s *bufio.Scanner) error

func (vimball *VimballArchive) Extract(folder string, r io.Reader, skipDir int) error {

	f := func(n string, l int, s *bufio.Scanner) error {
		return vimball.extractFile(n, l, s)
	}
	_, err := vimball.handle(folder, r, f)
	return err
}

func (vba *VimballArchive) Entries(r io.Reader, skipDir int) ([]string, error) {
	f := func(n string, l int, s *bufio.Scanner) error {
		return vba.skipFile(n, l, s)
	}
	return vba.handle("", r, f)
}

// extracts the contents a vimball formatted 'r' into 'dir'
// the format works like this:
//
// preamble
//
//	" Vimball Archive by Charles E. Campbell, Jr. Ph.D.
//	UseVimball
//	finish
//
// file-contents
//
//	folder/name_of_file
//	number_of_lines
//	...
//	...
//	folder2/other_file
//	number_of_lines2
//	...
//	...
func (vba *VimballArchive) handle(folder string, r io.Reader, extract vimballExtractFunc) ([]string, error) {

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	if err := vba.skipPreamble(scanner); err != nil {
		return nil, err
	}

	// now scan the file-entries
	contents := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.SplitN(line, "\t", 2)

		name := strings.TrimSpace(fields[0])
		name = filepath.Clean(name)

		errorMsg := "error vimball: while scanning line-number for %q: %v"
		if !scanner.Scan() {
			return nil, fmt.Errorf(errorMsg, name, scanner.Err())
		}

		nlines, err := strconv.Atoi(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf(errorMsg, name, err)
		}
		if nlines < 0 {
			return nil, fmt.Errorf(errorMsg, name, nlines)
		}

		if err = extract(filepath.Join(folder, name), nlines, scanner); err != nil {
			return nil, err
		}
		contents = append(contents, name)
	}

	return contents, scanner.Err()
}

func (*VimballArchive) skipPreamble(scanner *bufio.Scanner) error {

	useVimball := false
	finish := false

	// scan for lines 'UseVimball', followed by 'finish'
	for scanner.Scan() && !useVimball && !finish {
		line := scanner.Text()
		if !useVimball && line == "UseVimball" {
			useVimball = true
		} else if useVimball && !finish && line == "finish" {
			finish = true
		}
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	if !useVimball && !finish {
		return fmt.Errorf("error vimball: strange preamble")
	}

	return nil
}

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
		file.Write([]byte{'\n'})
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

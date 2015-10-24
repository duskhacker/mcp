package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var srcDir, dstDir string
var verbose, delete bool
var isImage *regexp.Regexp

func init() {
	flag.BoolVar(&verbose, "v", false, "Verbose")
	flag.BoolVar(&delete, "r", false, "Remove Source File")
	flag.StringVar(&srcDir, "s", "", "Source Directory")
	flag.StringVar(&dstDir, "d", "", "Destination Directory")
	isImage = regexp.MustCompile(`(?i)\.jpg$|\.mp4$|\.png$|\.avi$|\.mov$|\.tiff$`)
}

func main() {

	flag.Parse()

	if srcDir == "" || dstDir == "" {
		fmt.Fprintf(os.Stderr, "Usage: mcp [-v (verbose)] <-s src-dir> < -d dest-dir> -r (Remove)")
		os.Exit(-1)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Source: %s\nDestination: %s\n", srcDir, dstDir)
	}

	if _, err := os.Stat(dstDir); err != nil {
		fmt.Fprint(os.Stderr, "Destination Directory does not exist\n")
		os.Exit(-1)
	}

	filepath.Walk(srcDir, move)

	if verbose {
		fmt.Fprintf(os.Stderr, "Done")
	}

	os.Exit(0)
}

func move(srcFilename string, file os.FileInfo, err error) error {

	if !file.IsDir() && isImage.MatchString(file.Name()) {
		destFilename := filepath.Join(dstDir, file.Name())
		extension := filepath.Ext(file.Name())
		base := strings.TrimSuffix(file.Name(), extension)
		extension = strings.ToLower(extension)

		i := 0
		for {
			i++
			if _, err = os.Stat(destFilename); err == nil {
				destFilename = filepath.Join(dstDir, base+"-"+strconv.Itoa(i)+extension)
			} else {
				break
			}
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "Copying %s to %s\n", srcFilename, destFilename)
		}

		if err = cp(srcFilename, destFilename); err != nil {
			fmt.Fprintf(os.Stderr, "Error copying %s to %s: %s\n", srcFilename, destFilename, err)
			return err
		}

		if delete {
			fmt.Fprintf(os.Stderr, "Removing %s\n", srcFilename)
			if err = os.Remove(srcFilename); err != nil {
				fmt.Fprintf(os.Stderr, "Error removing %s: %s\n", srcFilename, err)
				return err
			}
		}
	}

	return nil
}

func cp(src string, dst string) error {
	var s, d *os.File
	var err error

	if s, err = os.Open(src); err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	if d, err = os.Create(dst); err != nil {
		return err
	}
	if _, err = io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

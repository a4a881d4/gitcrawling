package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ZipEncoder struct {
	archive *zip.Writer
	zipfile *os.File
}

func NewZipEnc(destZip string) (*ZipEncoder, error) {
	zipfile, err := os.Create(destZip)
	if err != nil {
		return nil, err
	}
	archive := zip.NewWriter(zipfile)
	return &ZipEncoder{
		archive: archive,
		zipfile: zipfile,
	}, nil
}

func (z *ZipEncoder) Close() {
	z.archive.Close()
	z.zipfile.Close()
}

// srcFile could be a single file or a directory
func (z *ZipEncoder) Zip(srcFile, owner string) error {
	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = path //owner + "/" + strings.TrimPrefix(path, filepath.Dir(srcFile)+"/")
		fmt.Println(header.Name)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := z.archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})

	return nil
}

func Zip() {
	z, err := NewZipEnc(flag.Arg(1))
	if err != nil {
		fmt.Println(err)
		return
	}

	var putSome = func(names []string) {
		for _, name := range names {
			repo := strings.Split(name, "/")
			if len(repo) != 2 {
				fmt.Println("error name", name)
				continue
			}
			owner, project := repo[0], repo[1]
			path := fmt.Sprintf("%s/repos/%s", *argReposDir, name)
			_, err := os.Stat(path)
			if err != nil {
				fmt.Println(ShowName(owner, project), "miss")
				missNum++
			} else {
				fmt.Println(ShowName(owner, project), "zip")
				err = z.Zip(path, owner)
			}
			repoNum++
		}
	}
	batchDo(putSome)
	fmt.Printf("%8d/%d\n", repoNum-missNum, repoNum)
	z.Close()
}

func Unzip() {
	err := unzip(flag.Arg(0), *argReposDir)
	if err != nil {
		fmt.Println(err)
	}
}

func unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		fpath = strings.Replace(fpath, `\`, `/`, -1)
		fmt.Println(fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
			outFile.Close()

		}
	}
	return nil
}

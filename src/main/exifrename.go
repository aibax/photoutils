package main

import (
	"flag"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"log"
	"os"
	"path/filepath"
)

const (
	/* ファイル名の日付時刻部分のフォーマット */
	DEFAULT_DATETIME_FORMAT string = "20060102_150405"

	/* カウンタ部分の桁数 */
	DEFAULT_COUNTER_LENGTH int = 2

	/* コマンドのヘルプメッセージで表示する文章 */
	USAGE = `
Usage of %s:
   %s [OPTIONS] FILES...
Options
`
)

func main() {

	/* Options */
	var prefix string
	var suffix string
	var datetimeFormat string = DEFAULT_DATETIME_FORMAT
	var counterLength int = DEFAULT_COUNTER_LENGTH
	var dryrun bool

	flag.StringVar(&prefix, "prefix", "", "ファイル名の先頭に付与するプレフィックスを入力します")
	flag.StringVar(&suffix, "suffix", "", "ファイル名の末尾に付与するサフィックスを入力します")
	flag.IntVar(&counterLength, "cl", DEFAULT_COUNTER_LENGTH, "カウンタ部分の桁数を入力します")
	flag.BoolVar(&dryrun, "dry-run", false, "実際に実行せずに実行結果を表示します")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, USAGE, os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	args := flag.Args()

	for _, filename := range args {
		_, err := rename(filename, prefix, suffix, datetimeFormat, counterLength, dryrun)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func rename(filename string, prefix string, suffix string, datetimeFormat string, counterLength int, dryrun bool) (result int, err error) {

	if !exists(filename) {
		result = 1
		err = os.ErrNotExist
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		result = 1
		return
	}

	e, err := exif.Decode(f)
	if err != nil {
		result = 1
		return
	}

	t, err := e.DateTime()
	datetime := t.Format(datetimeFormat)

	ext := filepath.Ext(filename)

	for i := 0; ; i++ {
		counter := fmt.Sprintf(fmt.Sprintf("_%%0%dd", counterLength), i)
		n := datetime + counter + ext

		if !exists(n) {
			newname := prefix + datetime + counter + suffix + ext
			log.Print("[RENAME] ", filename, " => ", newname)

			if !dryrun {
				err := os.Rename(filename, newname)
				if err != nil {
					log.Fatal(err)
				}
			}

			break
		}
	}

	result = 0
	return
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

const (
	// DefaultDatetimeFormat ファイル名の日付時刻部分の既定のフォーマット
	DefaultDatetimeFormat string = "20060102_150405"

	// DefaultCounterLength カウンタ部分の桁数の既定値
	DefaultCounterLength int = 2

	// Usage コマンドのヘルプメッセージで表示する文章
	Usage = `
Usage of %s:
   %s [OPTIONS] FILES...
Options
`
)

func main() {

	/* Options */
	var prefix string
	var suffix string
	var datetimeFormat = DefaultDatetimeFormat
	var counterLength = DefaultCounterLength
	var dryrun bool

	flag.StringVar(&prefix, "prefix", "", "ファイル名の先頭に付与するプレフィックスを入力します")
	flag.StringVar(&suffix, "suffix", "", "ファイル名の末尾に付与するサフィックスを入力します")
	flag.IntVar(&counterLength, "cl", DefaultCounterLength, "カウンタ部分の桁数を入力します")
	flag.BoolVar(&dryrun, "dry-run", false, "実際に実行せずに実行結果を表示します")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, Usage, os.Args[0], os.Args[0])
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
	if err != nil {
		if !exif.IsTagNotPresentError(err) {
			result = 1
			return
		}

		// Exifから撮影時刻が取得できない場合 => ファイルの更新時刻を使用
		info, _ := os.Stat(filename)
		t = info.ModTime()
		err = nil
	}

	datetime := t.Format(datetimeFormat)

	ext := filepath.Ext(filename)

	// ファイル名の抜けを防ぐため一時的なファイル名にリネームしてから処理を実行
	tempname := fmt.Sprint(time.Now().UnixNano()) + ext

	if !dryrun {
		err := os.Rename(filename, tempname)
		if err != nil {
			log.Fatal(err)
		}
	}

	for i := 0; ; i++ {
		counter := fmt.Sprintf(fmt.Sprintf("_%%0%dd", counterLength), i)
		n := datetime + counter + ext

		if !exists(n) {
			newname := prefix + datetime + counter + suffix + ext
			log.Print("[RENAME] ", filename, " => ", newname)

			if !dryrun {
				err := os.Rename(tempname, newname)
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

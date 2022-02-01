package loglog

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// accessErrorMessage は他アプリケーションからのアクセス中エラーメッセージ
const accessErrorMessage = "The process cannot access the file because it is being used by another process"

// interval はログファイルOPENのリトライ間隔
const interval time.Duration = 100 * time.Millisecond

// RetrySeconds はログファイルOPENのリトライ時間[秒]
var RetrySeconds int = 10

// IsAutoDelete はログの自動削除機能(ログ出力時に実行)を有効にするかどうか
var IsAutoDelete bool

// IsTrimNewLine は出力メッセージの改行コードを除去するかどうか
var IsTrimNewLine bool = true

// KeepDays はログ保持期間
var KeepDays = 14

// Directory はログ出力先フォルダ
var Directory = "log"

// lastDeleteDay は最後にログ削除関数を実行した日付(yyyyMMdd)
var lastDeleteDay string

// spaces はIsTrimNewLine=falseで出力するときのインデント用スペース文字列
var spaces = fmt.Sprintf("%"+strconv.Itoa(len("2006/01/02 15:01:02 : "))+"s", "")

const (
	CR   = "\r"
	LF   = "\n"
	CRLF = CR + LF
)

// Write はログ出力
func Write(message string) bool {
	now := time.Now()
	os.Mkdir(Directory, 0666)
	logname := filepath.Join(Directory, now.Format("20060102")+".log")

	today := now.Format("20060102")
	if IsAutoDelete && today != lastDeleteDay {
		Delete()
	}

retry:
	f, err := os.OpenFile(logname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		if strings.Contains(err.Error(), accessErrorMessage) {
			if int(time.Since(now).Seconds()) > RetrySeconds {
				return false
			}
			time.Sleep(interval)
			goto retry
		}
		return false
	}
	defer f.Close()

	write(f, message)
	return true
}

func write(f *os.File, message string) {
	if IsTrimNewLine {
		message = strings.Replace(message, LF, "", -1)
		message = strings.Replace(message, CR, "", -1)
	} else {
		var newMessage string
		msg := strings.Split(message, LF)
		for i, m := range msg {
			if i == 0 {
				newMessage += m
			} else {
				temp := strings.Replace(m, CR, "", -1)
				if temp != "" {
					newMessage += LF + spaces + temp
				}
			}
		}
		message = newMessage
	}
	logger := log.New(f, "", log.Ldate|log.Ltime)
	logger.Println(": " + message)
}

// Delete はログ削除
func Delete() bool {
	files, err := ioutil.ReadDir(Directory)
	if err != nil {
		return false
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	keepDate := today.AddDate(0, 0, 1-KeepDays)
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		tm, err := time.Parse("20060102", f.Name()[:len(f.Name())-4])
		if err != nil {
			continue
		}

		if tm.Before(keepDate) {
			os.Remove(filepath.Join(Directory, f.Name()))
		}
	}

	lastDeleteDay = now.Format("20060102")
	return true
}

package compare

import (
	"bufio"
	"fmt"
	"github.com/itnotebooks/fireye/config"
	"io"
	"log"
	"os"
	"strings"
)

func Analyze(appName, mtime, f string) string {

	file, err := os.Open(f)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer file.Close()

	exclude_keys := config.GLOBAL_CONFIG.ExcludeKeys

	// bufio 读取文件
	reader := bufio.NewReader(file)

	startKey := false
	var fileStr string

	for {
		str, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			return fileStr
		}

		if !startKey {
			// 比对日期
			if !strings.Contains(str, mtime) {
				log.Println("mtime:::",mtime)
				continue
			}
			// 比对行
			if !strings.Contains(str, config.GLOBAL_CONFIG.KeyWord) {
				log.Println("str:::",str)
				continue
			}

			startKey = true
		}
		fileStr += str
		log.Println(":::", fileStr)

		// 判断是否包含忽略关键字
		if len(exclude_keys) > 0 {
			for _, k := range exclude_keys {
				if strings.Contains(str, k) {
					fmt.Printf("%v: Skip the Keyword \"%v\" \n", appName, k)
					fileStr = ""
					startKey = false
					break
				}
			}

		}

	}
	log.Println(":::", fileStr)
	return fileStr
}

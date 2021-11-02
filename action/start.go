package action

import (
	"fmt"
	"github.com/itnotebooks/fireye/alert"
	"github.com/itnotebooks/fireye/compare"
	"github.com/itnotebooks/fireye/config"
	"github.com/itnotebooks/fireye/tools"
	"log"
	"os"
	"time"
)

func Factory(app config.LogDirField) {
	// 判断配置文件是否存在
	fileInfo, err := os.Stat(app.Path)
	if err != nil {
		fmt.Printf("-----------------------\n"+
			"%v\n"+
			"-----------------------\n", err)
		config.WG.Done()
		return
	}

	// 判断是否为目录
	if fileInfo.IsDir() {
		fmt.Printf("-----------------------\n"+
			"open %v: is a directory\n"+
			"-----------------------\n", app.Path)
		config.WG.Done()
		return
	}
	// 判断文件或目录是否存在
	if os.IsNotExist(err) {
		fmt.Printf("-----------------------\n"+
			"%v\n"+
			"-----------------------\n", err)
		config.WG.Done()
		return
	}

	// 判断最近的异常发生的时间是否在检测时间范围内
	errTimeCheck := compare.CompareTime(fileInfo.ModTime())
	if !errTimeCheck {
		fmt.Printf("-----------------------\n"+
			"Skip Over: %v\n"+
			"-----------------------\n", app.Path)
		config.WG.Done()
		return
	}

	fmt.Printf("Hit %v\n", app.Path)
	errMsg := compare.Analyze(app.Name, fileInfo.ModTime().Format(config.GLOBAL_CONFIG.DateFormat), app.Path)
	log.Println(errMsg)
	if errMsg == "" {
		config.WG.Done()
		return
	}
	// 发送钉钉通知
	if app.DingTalk != "" {
		config.WG.Add(1)
		go alert.SendDingTalkRobot(app, errMsg)
	}

	// 发送邮件通知
	if config.GLOBAL_CONFIG.SMTP.SMTP_ENABLE {
		config.WG.Add(1)
		go alert.SendMail(app, errMsg)
	}

	config.WG.Done()

}

func Start() {

	// 获取本机公网IP
	config.GLOBAL_CONFIG.GIP = tools.HttpGet(config.GLOBAL_CONFIG.GipCheck)
	config.GLOBAL_CONFIG.PIP = tools.GetOutBoundIP()

	beginTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\n==========%v==========\n", beginTime)

	for _, app := range config.GLOBAL_CONFIG.LogDirs {
		config.WG.Add(1)
		go Factory(app)

	}

	config.WG.Wait()

	endTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\n==========%v==========\n", endTime)
}

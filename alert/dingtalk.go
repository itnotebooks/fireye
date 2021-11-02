package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/itnotebooks/fireye/config"
	"io/ioutil"
	"net/http"
	"time"
)

// SendDingTalkRobot 发送错误日志到钉钉群
func SendDingTalkRobot(app config.LogDirField, err_msg string) {
	global_config := config.GLOBAL_CONFIG

	privateIP := global_config.PIP
	path := app.Path

	if global_config.Platform == "container" {
		privateIP = app.PodName
	}

	title := fmt.Sprintf("ERROR: %v webapps:%v (%v)", global_config.Project, app.Name, time.Now().Format("2006-01-02 15:04:05"))
	var err error
	data := make(map[string]interface{})

	if len(err_msg) > 15000 {
		err_msg = err_msg[:15000]
	}

	data["msgtype"] = "markdown"
	data["markdown"] = map[string]interface{}{
		"title": title,
		"text": fmt.Sprintf("## %v\n\n"+
			"-----\n\n"+
			"> **project:** %v\n\n"+
			"> **Mode:** %v\n\n"+
			"> **Global IP:** %v\n\n"+
			"> **Private IP:** %v\n\n"+
			"> **Log File:** %v\n\n"+
			"```\n\n"+
			"%v\n\n"+
			"```\n\n"+
			"**注:** 详情请查收邮件\n\n", title, global_config.Project, app.Name,
			global_config.GIP, privateIP, path, err_msg),
	}

	bytesData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		config.WG.Done()
		return
	}

	resp, err := http.Post(fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%v", app.DingTalk), "application/json", bytes.NewReader(bytesData))

	if err != nil {
		fmt.Println(err)
		config.WG.Done()
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		config.WG.Done()
		return
	}

	fmt.Printf("%v: DingTalk send success..., msg: %v\n", app.Name, string(body))

	config.WG.Done()
}

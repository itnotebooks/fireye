package config

import (
	"flag"
	"fmt"
	"github.com/itnotebooks/fireye/tools/k8s"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

type SmtpField struct {
	SMTP_ENABLE   bool   `yaml:"smtp_enable"`
	SMTP_SERVER   string `yaml:"smtp_address"`
	SMTP_PORT     int    `yaml:"smtp_port"`
	SMTP_USERNAME string `yaml:"smtp_username"`
	SMTP_PASSWORD string `yaml:"smtp_password"`
	SMTP_STARTTLS bool   `yaml:"smtp_starttls"`
}

type LogDirField struct {
	Name     string `yaml:"name" json:"name"`
	PodName  string `yaml:"pod_name" json:"pod_name"`
	Path     string `yaml:"path" json:"path"`
	RealPath string `yaml:"real_path" json:"real_path"`
	DingTalk string `yaml:"dingtalk_accesstoken" json:"dingtalk_accesstoken"`
}

type DebugField struct {
	ENABLE   bool     `yaml:"enable" bson:"enable" json:"enable"`
	Minutes  int      `yaml:"minutes" json:"minutes"`
	DingTalk string   `yaml:"dingtalk_accesstoken" json:"dingtalk_accesstoken"`
	MailTo   []string `yaml:"mail_to"  json:"mail_to"`
}

type ConfigField struct {
	Project    string        `yaml:"project" json:"project"`
	Department string        `yaml:"department" json:"department"`
	Platform   string        `yaml:"platform" json:"platform"`   // 运行平台，值为container或其它，当为container时表示运行在容器环境
	NameSpace  string        `yaml:"namespace" json:"namespace"` // 当platform为container时生效，表示应用运行的命名空间
	LogDirs    []LogDirField `yaml:"logdirs" json:"logdirs"`     // 当platform为container时，此项可为空，表示采集目标namespace
	// 下所有deployment
	Minutes     int        `yaml:"minutes" json:"minutes"`
	DingTalk    string     `yaml:"dingtalk_accesstoken" json:"dingtalk_accesstoken"`
	SMTP        SmtpField  `yaml:"smtp" json:"smtp"`
	Log         string     `yaml:"log" json:"log"`
	GipCheck    string     `yaml:"gip_check" json:"gip_check"`
	MailTo      []string   `yaml:"mail_to"  json:"mail_to"`
	MailCC      []string   `yaml:"mail_cc" json:"mail_cc"`
	ExcludeKeys []string   `yaml:"exclude_keys" json:"exclude_keys"`
	DEBUG       DebugField `json:"debug" yaml:"debug"`
	KeyWord     string     `yaml:"keyWord" json:"keyWord"`
	DateFormat  string     `yaml:"dateFormat" json:"dateFormat"`
	GIP         string
	PIP         string
}

var WG sync.WaitGroup

// 用于存放全局配置信息
var GLOBAL_CONFIG ConfigField

func RenderConfig(c string) {

	var err error
	// c := ConfigField{}
	var config ConfigField

	// 判断是否为文件并读取文件内容
	f, err := ioutil.ReadFile(c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)

	}

	err = yaml.Unmarshal(f, &config)
	if err != nil {
		fmt.Println(err)
	}

	// 判断项目名字段是否为空，如果为空则直接退出
	if strings.Trim(config.Project, " ") == "" {
		fmt.Println("项目名不能为空")
		os.Exit(1)
	}

	GLOBAL_CONFIG.Project = strings.Trim(config.Project, " ")
	GLOBAL_CONFIG.DingTalk = strings.Trim(config.DingTalk, " ")
	GLOBAL_CONFIG.GipCheck = strings.Trim(config.GipCheck, " ")
	GLOBAL_CONFIG.Minutes = config.Minutes
	GLOBAL_CONFIG.MailTo = config.MailTo
	GLOBAL_CONFIG.MailCC = config.MailCC

	// 关键关定义，默认：ERROR
	GLOBAL_CONFIG.KeyWord = config.KeyWord
	if config.KeyWord == "" {
		GLOBAL_CONFIG.KeyWord = "ERROR"
	}
	// 日期格式赋值
	GLOBAL_CONFIG.DateFormat = config.DateFormat
	if config.DateFormat == "" {
		GLOBAL_CONFIG.DateFormat = "2006-01-02 15:04"
	}

	// 调试模式
	if config.DEBUG.ENABLE {
		GLOBAL_CONFIG.DEBUG = config.DEBUG
		GLOBAL_CONFIG.Minutes = config.DEBUG.Minutes
		GLOBAL_CONFIG.MailTo = config.DEBUG.MailTo
		GLOBAL_CONFIG.MailCC = make([]string, 0)
		GLOBAL_CONFIG.DingTalk = config.DEBUG.DingTalk
	}

	if strings.Trim(config.Platform, " ") == "container" {
		NameSpace := strings.Trim(config.NameSpace, " ")

		if NameSpace == "" {
			NameSpace = "default"
		}

		GLOBAL_CONFIG.NameSpace = NameSpace
		GLOBAL_CONFIG.Platform = "container"
		client := k8s.NewK8SClient()
		// 获取所有Deployment
		for _, deployment := range client.GetDeploymentList(NameSpace) {
			// 获取Deployment下所有Pod，并按规则生产Pod的日志路径
			for _, pod := range client.GetPodList(NameSpace, deployment) {
				deploymentName := deployment
				GLOBAL_CONFIG.LogDirs = append(GLOBAL_CONFIG.LogDirs, LogDirField{
					Name:     deploymentName,
					PodName:  pod,
					Path:     fmt.Sprintf("/var/apps/logs/%s/%s.%s.log.error", deploymentName, deploymentName, pod),
					DingTalk: GLOBAL_CONFIG.DingTalk,
				})
			}
		}
	} else {
		if len(config.LogDirs) < 1 {
			fmt.Println("Error日志路径只少有一个")
			os.Exit(1)
		}

		// 获取模块及对应的Error日志路径配置信息
		for _, v := range config.LogDirs {
			loginfo := LogDirField{}

			name := strings.Trim(v.Name, " ")
			path := strings.Trim(v.Path, " ")
			dingtalk := strings.Trim(v.DingTalk, " ")

			if name == "" || path == "" {
				fmt.Println("日志模块名或日志路径不能为空")
				continue
			}

			// 如果没有为模块配置独立钉钉通知群，则使用全局配置
			if dingtalk == "" {
				dingtalk = GLOBAL_CONFIG.DingTalk
			}

			loginfo.Name = name
			loginfo.Path = path
			// 调试模式
			if config.DEBUG.ENABLE {
				loginfo.DingTalk = GLOBAL_CONFIG.DingTalk
			} else {
				loginfo.DingTalk = dingtalk
			}
			GLOBAL_CONFIG.LogDirs = append(GLOBAL_CONFIG.LogDirs, loginfo)

		}

		if len(GLOBAL_CONFIG.LogDirs) < 1 {
			fmt.Println(GLOBAL_CONFIG.LogDirs)
			os.Exit(1)
		}
	}

	// SMTP配置
	if config.SMTP.SMTP_ENABLE {
		GLOBAL_CONFIG.SMTP = config.SMTP
		GLOBAL_CONFIG.SMTP.SMTP_ENABLE = true

	}

	if GLOBAL_CONFIG.GipCheck == "" {
		GLOBAL_CONFIG.GipCheck = "http://ip.dhcp.cn/?ip"
	}

	// 忽略关键字
	GLOBAL_CONFIG.ExcludeKeys = config.ExcludeKeys

}

// 读取配置文件
func GetConfig() string {
	var config string
	// 获取配置文件目录
	flag.StringVar(&config, "c", "./config.yaml", "配置文件")

	flag.Parse()

	// 判断配置文件是否存在
	_, err := os.Stat(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// 判断文件或目录是否存在
	if os.IsNotExist(err) {
		fmt.Println(err)
		os.Exit(1)
	}

	return config
}

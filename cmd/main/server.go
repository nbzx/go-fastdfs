package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nbzx/go-fastdfs/server"
	"github.com/sjqzhang/goutil"
	log "github.com/sjqzhang/seelog"
)

type Server struct {
	server.Server
}

var (
	logacc         log.LoggerInterface
	staticHandler  http.Handler
	serverInstance *server.Server = nil
	logConfigStr                  = `
<seelog type="asynctimer" asyncinterval="1000" minlevel="trace" maxlevel="error">  
	<outputs formatid="common">  
		<buffered formatid="common" size="1048576" flushperiod="1000">  
			<rollingfile type="size" filename="{DOCKER_DIR}log/fileserver.log" maxsize="104857600" maxrolls="10"/>  
		</buffered>
	</outputs>  	  
	 <formats>
		 <format id="common" format="%Date %Time [%LEV] [%File:%Line] [%Func] %Msg%n" />  
	 </formats>  
</seelog>
`
	logAccessConfigStr = `
<seelog type="asynctimer" asyncinterval="1000" minlevel="trace" maxlevel="error">  
	<outputs formatid="common">  
		<buffered formatid="common" size="1048576" flushperiod="1000">  
			<rollingfile type="size" filename="{DOCKER_DIR}log/access.log" maxsize="104857600" maxrolls="10"/>  
		</buffered>
	</outputs>  	  
	 <formats>
		 <format id="common" format="%Date %Time [%LEV] [%File:%Line] [%Func] %Msg%n" />  
	 </formats>  
</seelog>
`
	util = &goutil.Common{}
)

const (
	cfgJson = `{
		"绑定端号": "端口",
		"addr": ":8080",
		"是否开启https": "默认不开启，如需启开启，请在conf目录中增加证书文件 server.crt 私钥 文件 server.key",
		"enable_https": false,
		"PeerID": "集群内唯一,请使用0-9的单字符，默认自动生成",
		"peer_id": "%s",
		"本主机地址": "本机http地址,默认自动生成(注意端口必须与addr中的端口一致），必段为内网，自动生成不为内网请自行修改，下同",
		"host": "%s",
		"集群": "集群列表,注意为了高可用，IP必须不能是同一个,同一不会自动备份，且不能为127.0.0.1,且必须为内网IP，默认自动生成",
		"peers": ["%s"],
		"组号": "用于区别不同的集群(上传或下载)与support_group_manage配合使用,带在下载路径中",
		"group": "group1",
		"是否支持按组（集群）管理,主要用途是Nginx支持多集群": "默认支持,不支持时路径为http://10.1.5.4:8080/action,支持时为http://10.1.5.4:8080/group(配置中的group参数)/action,action为动作名，如status,delete,sync等",
		"support_group_manage": true,
		"是否合并小文件": "默认不合并,合并可以解决inode不够用的情况（当前对于小于1M文件）进行合并",
		"enable_merge_small_file": false,
		"图片是否缩放": "默认是",
		"enable_image_resize": true,
		"图片最大宽度": "默认值2000",
		"image_max_width": 2000,
		"图片最大高度": "默认值1000",
		"image_max_height": 1000,
		"允许后缀名": "允许可以上传的文件后缀名，如jpg,jpeg,png等。留空允许所有。",
		"extensions": [],
		"重试同步失败文件的时间": "单位秒",
		"refresh_interval": 1800,
		"是否自动重命名": "默认不自动重命名,使用原文件名",
		"rename_file": false,
		"是否支持web上传,方便调试": "默认支持web上传",
		"enable_web_upload": true,
		"enable_pprof_debug": false,
		"是否支持非日期路径": "默认支持非日期路径,也即支持自定义路径,需要上传文件时指定path",
		"enable_custom_path": true,
		"下载域名": "用于外网下载文件的域名",
		"download_domain": "",
		"场景列表": "当设定后，用户指的场景必项在列表中，默认不做限制(注意：如果想开启场景认功能，格式如下：'场景名:googleauth_secret' 如 default:N7IET373HB2C5M6D ",
		"scenes": [],
		"默认场景": "默认default",
		"default_scene": "default",
		"是否显示目录": "默认显示,方便调试用,上线时请关闭",
		"show_dir": true,
		"邮件配置": "",
		"mail": {
			"user": "abc@163.com",
			"password": "abc",
			"host": "smtp.163.com:25"
		},
		"反向代理缓存内容":"目前只支持pypi ex:  pip install -i http://127.0.0.1:9000/simple pandas",
		"proxies":[
			{"dir":"pypi","origin":"https://pypi.douban.com","addr":":9000"}
		],
		"告警接收邮件列表": "接收人数组",
		"alarm_receivers": [],
		"告警接收URL": "方法post,参数:subject,message",
		"alarm_url": "",
		"下载是否需带token": "真假",
		"download_use_token": false,
		"下载token过期时间": "单位秒",
		"download_token_expire": 600,
		"是否自动修复": "在超过1亿文件时出现性能问题，取消此选项，请手动按天同步，请查看FAQ",
		"auto_repair": true,
		"文件去重算法md5可能存在冲突，默认md5": "sha1|md5",
		"file_sum_arithmetic": "md5",
		"管理ip列表": "用于管理集的ip白名单,如果放开所有内网则可以用 0.0.0.0 ,注意为了安全，不对外网开放",
		"admin_ips": ["127.0.0.1"],
		"是否启用迁移": "默认不启用",
		"enable_migrate": false,
		"文件是否去重": "默认去重",
		"enable_distinct_file": true,
		"是否开启跨站访问": "默认开启",
		"enable_cross_origin": true,
		"是否开启Google认证，实现安全的上传、下载": "默认不开启",
		"enable_google_auth": false,
		"认证url": "当url不为空时生效,注意:普通上传中使用http参数 auth_token 作为认证参数, 在断点续传中通过HTTP头Upload-Metadata中的auth_token作为认证参数,认证流程参考认证架构图",
		"auth_url": "",
		"下载是否认证": "默认不认证(注意此选项是在auth_url不为空的情况下生效)",
		"enable_download_auth": false,
		"默认是否下载": "默认下载",
		"default_download": true,
		"本机是否只读": "默认可读可写",
		"read_only": false,
		"是否开启断点续传": "默认开启",
		"enable_tus": true,
		"同步单一文件超时时间（单位秒）": "默认为0,程序自动计算，在特殊情况下，自已设定",
		"sync_timeout": 0
	}
		`
)

func InitServer(appDir string) {
	server.DOCKER_DIR = os.Getenv("GO_FASTDFS_DIR")
	if server.DOCKER_DIR == "" {
		server.DOCKER_DIR = appDir
	}
	if server.DOCKER_DIR != "" {
		if !strings.HasSuffix(server.DOCKER_DIR, "/") {
			server.DOCKER_DIR = server.DOCKER_DIR + "/"
		}
	}
	server.STORE_DIR = server.DOCKER_DIR + server.STORE_DIR_NAME
	server.CONF_DIR = server.DOCKER_DIR + server.CONF_DIR_NAME
	server.DATA_DIR = server.DOCKER_DIR + server.DATA_DIR_NAME
	server.LOG_DIR = server.DOCKER_DIR + server.LOG_DIR_NAME
	server.STATIC_DIR = server.DOCKER_DIR + server.STATIC_DIR_NAME
	server.LARGE_DIR_NAME = "haystack"
	server.LARGE_DIR = server.STORE_DIR + "/haystack"
	server.CONST_LEVELDB_FILE_NAME = server.DATA_DIR + "/fileserver.db"
	server.CONST_LOG_LEVELDB_FILE_NAME = server.DATA_DIR + "/log.db"
	server.CONST_STAT_FILE_NAME = server.DATA_DIR + "/stat.json"
	server.CONST_CONF_FILE_NAME = server.CONF_DIR + "/cfg.json"
	server.CONST_SERVER_CRT_FILE_NAME = server.CONF_DIR + "/server.crt"
	server.CONST_SERVER_KEY_FILE_NAME = server.CONF_DIR + "/server.key"
	server.CONST_SEARCH_FILE_NAME = server.DATA_DIR + "/search.txt"
	server.FOLDERS = []string{server.DATA_DIR, server.STORE_DIR, server.CONF_DIR, server.STATIC_DIR}
	logAccessConfigStr = strings.Replace(logAccessConfigStr, "{DOCKER_DIR}", server.DOCKER_DIR, -1)
	logConfigStr = strings.Replace(logConfigStr, "{DOCKER_DIR}", server.DOCKER_DIR, -1)
	for _, folder := range server.FOLDERS {
		os.MkdirAll(folder, 0775)
	}
	serverInstance = server.NewServer()

	var peerId string
	if peerId = os.Getenv("GO_FASTDFS_PEER_ID"); peerId == "" {
		peerId = fmt.Sprintf("%d", util.RandInt(0, 9))
	}
	if !util.FileExists(server.CONST_CONF_FILE_NAME) {
		var ip string
		if ip = os.Getenv("GO_FASTDFS_IP"); ip == "" {
			ip = util.GetPulicIP()
		}
		peer := "http://" + ip + ":8080"
		var peers string
		if peers = os.Getenv("GO_FASTDFS_PEERS"); peers == "" {
			peers = peer
		}
		cfg := fmt.Sprintf(cfgJson, peerId, peer, peers)
		util.WriteFile(server.CONST_CONF_FILE_NAME, cfg)
	}
	if logger, err := log.LoggerFromConfigAsBytes([]byte(logConfigStr)); err != nil {
		panic(err)
	} else {
		log.ReplaceLogger(logger)
	}
	if _logacc, err := log.LoggerFromConfigAsBytes([]byte(logAccessConfigStr)); err == nil {
		logacc = _logacc
		log.Info("succes init log access")
	} else {
		log.Error(err.Error())
	}
	server.ParseConfig(server.CONST_CONF_FILE_NAME)
	if ips, _ := util.GetAllIpsV4(); len(ips) > 0 {
		_ip := util.Match("\\d+\\.\\d+\\.\\d+\\.\\d+", server.Config().Host)
		if len(_ip) > 0 && !util.Contains(_ip[0], ips) {
			msg := fmt.Sprintf("host config is error,must in local ips:%s", strings.Join(ips, ","))
			log.Warn(msg)
			fmt.Println(msg)
		}
	}
	if server.Config().QueueSize == 0 {
		server.Config().QueueSize = server.CONST_QUEUE_SIZE
	}
	if server.Config().PeerId == "" {
		server.Config().PeerId = peerId
	}
	if server.Config().SupportGroupManage {
		staticHandler = http.StripPrefix("/"+server.Config().Group+"/", http.FileServer(http.Dir(server.STORE_DIR)))
	} else {
		staticHandler = http.StripPrefix("/", http.FileServer(http.Dir(server.STORE_DIR)))
	}
	serverInstance.InitComponent(false)
}

func Start() {
	serverInstance.Start()
}

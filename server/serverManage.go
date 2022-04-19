package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/sjqzhang/seelog"
)

var (
	cfgJsonPort = strings.Replace(cfgJson, ":8080", "%s", 1)
)

func ConfigServer(defaultHttpAddr, appDir string) *Server {
	DOCKER_DIR = os.Getenv("GO_FASTDFS_DIR")
	if DOCKER_DIR == "" {
		DOCKER_DIR = appDir
	}
	if DOCKER_DIR != "" {
		if !strings.HasSuffix(DOCKER_DIR, "/") {
			DOCKER_DIR = DOCKER_DIR + "/"
		}
	}
	STORE_DIR = DOCKER_DIR + STORE_DIR_NAME
	CONF_DIR = DOCKER_DIR + CONF_DIR_NAME
	DATA_DIR = DOCKER_DIR + DATA_DIR_NAME
	LOG_DIR = DOCKER_DIR + LOG_DIR_NAME
	STATIC_DIR = DOCKER_DIR + STATIC_DIR_NAME
	LARGE_DIR_NAME = "haystack"
	LARGE_DIR = STORE_DIR + "/haystack"
	CONST_LEVELDB_FILE_NAME = DATA_DIR + "/filedb"
	CONST_LOG_LEVELDB_FILE_NAME = DATA_DIR + "/log.db"
	CONST_STAT_FILE_NAME = DATA_DIR + "/stat.json"
	CONST_CONF_FILE_NAME = CONF_DIR + "/cfg.json"
	CONST_SERVER_CRT_FILE_NAME = CONF_DIR + "/crt"
	CONST_SERVER_KEY_FILE_NAME = CONF_DIR + "/key"
	CONST_SEARCH_FILE_NAME = DATA_DIR + "/search.txt"
	FOLDERS = []string{DATA_DIR, STORE_DIR, CONF_DIR, STATIC_DIR}
	logAccessConfigStr = strings.Replace(logAccessConfigStr, "{DOCKER_DIR}", DOCKER_DIR, -1)
	logConfigStr = strings.Replace(logConfigStr, "{DOCKER_DIR}", DOCKER_DIR, -1)
	for _, folder := range FOLDERS {
		os.MkdirAll(folder, 0775)
	}
	server = NewServer()

	var peerId string
	if peerId = os.Getenv("GO_FASTDFS_PEER_ID"); peerId == "" {
		peerId = fmt.Sprintf("%d", server.util.RandInt(0, 9))
	}
	if !server.util.FileExists(CONST_CONF_FILE_NAME) {
		var ip string
		if ip = os.Getenv("GO_FASTDFS_IP"); ip == "" {
			ip = server.util.GetPulicIP()
		}
		peer := "http://" + ip + ":8080"
		var peers string
		if peers = os.Getenv("GO_FASTDFS_PEERS"); peers == "" {
			peers = peer
		}
		cfg := fmt.Sprintf(cfgJsonPort, defaultHttpAddr, peerId, peer, peers)
		server.util.WriteFile(CONST_CONF_FILE_NAME, cfg)
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
	ParseConfig(CONST_CONF_FILE_NAME)
	if ips, _ := server.util.GetAllIpsV4(); len(ips) > 0 {
		_ip := server.util.Match("\\d+\\.\\d+\\.\\d+\\.\\d+", Config().Host)
		if len(_ip) > 0 && !server.util.Contains(_ip[0], ips) {
			msg := fmt.Sprintf("host config is error,must in local ips:%s", strings.Join(ips, ","))
			log.Warn(msg)
			fmt.Println(msg)
		}
	}
	if Config().QueueSize == 0 {
		Config().QueueSize = CONST_QUEUE_SIZE
	}
	if Config().PeerId == "" {
		Config().PeerId = peerId
	}
	if Config().SupportGroupManage {
		staticHandler = http.StripPrefix("/"+Config().Group+"/", http.FileServer(http.Dir(STORE_DIR)))
	} else {
		staticHandler = http.StripPrefix("/", http.FileServer(http.Dir(STORE_DIR)))
	}
	server.initComponent(false)
	return server
}

func StartServer(ctx context.Context) {
	server.Start(ctx)
}

package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
)

//基于协程的tcp压测工具
type Config struct {
	TargetIp     string `yaml:"TargetIp"`
	TargetPort   int    `yaml:"TargetPort"`
	CoroutineNum int    `yaml:"CoroutineNum"`
	Duration     int    `yaml:"Duration"`
	MaxTimeout   int    `yaml:"MaxTimeout"`
	Cmd          string `yaml:"Cmd"`
	// 在这里添加其他需要读取的字段
}

var config Config

func init() {
	logFile, err := os.OpenFile("./wrk_tcp_tool.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetPrefix("[wrk_tcp_tool]")
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)

	pid := os.Getpid()
	log.Printf("wrk_tcp_tool start run in pid:%d,log setup ok", pid)
	// 读取YAML文件
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("config yaml read err!")
		panic(err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Printf("config yaml parse err!")
		panic(err)
	}
}

func main() {
	pid := os.Getpid()
	log.Printf("-----wrk_tcp_tool start run in pid:%d-----", pid)
	//压测目标接口  Ip port str 并发数,时长,
	//统计指标  总请求量,平均响应时长
	server := NewServer(config.TargetIp, config.TargetPort, config.CoroutineNum, config.Duration, config.MaxTimeout, config.Cmd)

	server.MutipleStart()
	log.Printf("-----wrk_tcp_tool start run over:%d-----", pid)
}

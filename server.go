package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	CoroutineNum int
	Duration     int
	MaxTimeout   int
	Cmd          string
}

func NewServer(Ip string, Port int, CoroutineNum int, Duration int, MaxTimeout int, Cmd string) *Server {
	server := &Server{
		Ip:           Ip,
		Port:         Port,
		CoroutineNum: CoroutineNum,
		Duration:     Duration,
		MaxTimeout:   MaxTimeout,
		Cmd:          Cmd,
	}
	return server
}

func (this *Server) RspHandler(Recv []byte) {
	failCount := 0
	successCount := 0
	recvStr := string(Recv)
	//log.Printf("handel data:%s", recvStr)
	parts := strings.Split(recvStr, "&")
	m := make(map[string]string)
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	result := m["result"]
	//content := m["content"]
	if result != "0" {
		failCount += 1
		log.Printf("fail rsp result:%s", result)
	} else {
		successCount += 1
	}

}

// Start this 是server的抽象
func (this *Server) Start() {
	pid := os.Getpid()
	log.Printf("-----wrk_tcp_tool run in pid:%d ok-----", pid)
	ipPort := this.Ip + ":" + strconv.Itoa(this.Port)
	log.Printf("-----wrk_tcp_tool start run in setting:%s-----", ipPort)
	var wg sync.WaitGroup
	duration := 2 * time.Second
	for i := 0; i < this.CoroutineNum; i++ {
		log.Printf("----try begin coroutine:%d-----", i)
		go func() {
			sTime := time.Now()
			wg.Add(1)
			for {
				log.Printf("----try dial:%s-----", ipPort)
				conn, err := net.Dial("tcp", ipPort)
				if err != nil {
					log.Println("Error connecting:", err)
					continue
				}
				_, err = conn.Write([]byte(this.Cmd + "\r\n"))
				if err != nil {
					log.Println("Error writing:", err)
					continue
				}
				var recvData []byte
				for {
					readBuf := make([]byte, 1024)
					recvNum, err := conn.Read(readBuf)
					if recvNum > 0 && err != nil && err != io.EOF {
						if bytes.Equal(readBuf[recvNum-2:recvNum], []byte{'\r', '\n'}) {
							recvData = append(recvData, readBuf...)
							break //接收数据完成
						}
						log.Println(recvNum)
					}
					if err != nil {
						log.Println("Error reading:", err)
					}
					recvData = append(recvData, readBuf...)
				}
				this.RspHandler(recvData)
				err = conn.Close()
				if err != nil {
					log.Println("Error close socket:", err)
					return
				}

				if time.Since(sTime) >= duration {
					log.Println("Time over!")
					break
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

}

func (this *Server) SingleStart() (time.Duration, []byte) {
	ipPort := this.Ip + ":" + strconv.Itoa(this.Port)
	//log.Printf("----try dial:%s-----", ipPort)

	conn, err := net.Dial("tcp", ipPort)
	if err != nil {
		log.Println("Error connecting:", err)
	}
	_, err = conn.Write([]byte(this.Cmd + "\r\n"))
	if err != nil {
		log.Println("Error writing:", err)
	}
	var recvData []byte
	readBuf := make([]byte, 1024)
	start := time.Now()
	for {
		recvNum, err := conn.Read(readBuf)
		if recvNum > 0 {
			if bytes.Equal(readBuf[recvNum-2:recvNum], []byte{'\r', '\n'}) {
				recvData = append(recvData, readBuf...)
				//log.Printf("reve end sign")
				break
			}
			if err != nil {
				log.Println("Error reading:", err)
			}
			recvData = append(recvData, readBuf...)
		}
	}
	elapsed := time.Since(start)
	//log.Printf("single req:%dms", int(elapsed.Milliseconds()))

	err = conn.Close()
	if err != nil {
		log.Println("Error close socket:", err)
		return 0, nil
	}
	return elapsed, readBuf
}
func (this *Server) MutipleStart() {
	pid := os.Getpid()
	log.Printf("-----wrk_tcp_tool run in pid:%d ok-----", pid)
	ipPort := this.Ip + ":" + strconv.Itoa(this.Port)
	log.Printf("-----wrk_tcp_tool start run in setting:%s-----", ipPort)
	var wg sync.WaitGroup
	var totalElapsed time.Duration
	reqCountRes := 0
	reqCount100ms := 0
	reqCount200ms := 0
	reqCount300ms := 0
	//reqCount500ms := 0

	sTime := time.Now()
	log.Printf("start time: %s", sTime)
	for i := 0; i < this.CoroutineNum; i++ {
		//log.Printf("----try begin coroutine:%d-----", i)
		wg.Add(1) //gon func外添加,保证goroutine在组内
		go func() {
			reqCount := 0
			for {
				if time.Since(sTime) > time.Second*time.Duration(this.Duration) {
					//log.Printf("Time over! %s", time.Now())
					break
				}
				reqCount += 1
				rspDuration, readBuf := this.SingleStart()
				totalElapsed += rspDuration

				if 0*time.Millisecond < rspDuration && rspDuration < 100*time.Millisecond {
					reqCount100ms += 1
				} else if 100*time.Millisecond <= rspDuration && rspDuration < 200*time.Millisecond {
					reqCount200ms += 1
				} else if 200*time.Millisecond <= rspDuration && rspDuration < 300*time.Millisecond {
					reqCount300ms += 1
				}
				this.RspHandler(readBuf)
			}
			reqCountRes += reqCount
			wg.Done()

		}()
	}
	wg.Wait()
	log.Printf("----total req in %d sencond %d-----", this.Duration, reqCountRes)
	log.Printf("QPms:%d/ms", reqCountRes/(this.Duration*1000))
	log.Printf("QPS_out:%d/s", reqCountRes/this.Duration)
	log.Printf("TOTAL_REQ:%d", reqCountRes)
	log.Printf("TOTAL_AVG:%dms", int(totalElapsed.Milliseconds()))
	log.Printf("<100_RSP:%.5f%%,num:%d", float64(reqCount100ms)/float64(reqCountRes)*100, reqCount100ms)
	log.Printf("<200_RSP:%.5f%%,num:%d", float64(reqCount200ms)/float64(reqCountRes)*100, reqCount200ms)
	log.Printf("<300_RSP:%.5f%%,num:%d", float64(reqCount300ms)/float64(reqCountRes)*100, reqCount300ms)
	log.Printf("AVG_RSP:%dms", int(totalElapsed.Milliseconds())/reqCountRes)
	log.Printf("QPS_RSP:%d/s", reqCountRes/int(totalElapsed.Seconds()))
	log.Printf("CoroutineNum:%d", this.CoroutineNum)
}

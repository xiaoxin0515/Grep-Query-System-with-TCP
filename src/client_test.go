package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type ServerInfo struct {
	ServerId  int
	ServerAdr string
}

type Grepres struct {
	Connected   bool
	LogfileName string
	MatchOrnot  bool
	MatchCount  int
}

type ClientRequest struct {
	Greppattern string
	LogFile     string
}

type TestInfo struct {
	GrepPatterm string
	LineCnt     int
}

type MatchRes struct {
	TotalLineCnt int
	TotalFileCnt int
}

func get_server_info(servers_adr string) []ServerInfo {
	file_adr, err := os.Open(servers_adr)
	if err != nil {
		log.Fatal("error with reading servers address file!", err)
	}
	defer file_adr.Close()
	buf := bufio.NewReader(file_adr)
	server_id := 0
	servers_info := []ServerInfo{}
	for true {
		address, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal("error with reading servers address file!", err)
			}
		}
		server_id += 1
		//fmt.Print(address)
		address = strings.TrimSpace(address)
		server := ServerInfo{server_id, address}
		servers_info = append(servers_info, server)
	}
	return servers_info
}
func TestGrepMatch(t *testing.T) {
	tests := get_test_info("test.txt")
	for _, test := range tests {
		res := grep_match(test.GrepPatterm, true)
		if res.TotalFileCnt == 1 {
			fmt.Println("1")
		}
		if res.TotalLineCnt == test.LineCnt {
			fmt.Println("line count match expected count")
		} else {
			fmt.Println("line count does not match expected count")
		}
		//fmt.Println(res.TotalFileCnt)
	}
}
func get_test_info(test_file string) []TestInfo {
	file_test, err := os.Open(test_file)
	if err != nil {
		log.Fatal("error with reading test file!", err)
	}
	defer file_test.Close()
	buf := bufio.NewReader(file_test)
	tests := []TestInfo{}
	for true {
		test, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal("error with reading test file!", err)
			}
		}
		//fmt.Print(test)
		//fmt.Println(strings.Split(test, " "))
		piece := strings.Split(test, " ")
		line_cnt, _ := strconv.Atoi(piece[1])
		test_piece := TestInfo{piece[0], line_cnt}
		tests = append(tests, test_piece)
	}
	return tests
}
func grep_match(reg_pattern string, test bool) MatchRes {
	servers_adr := "/home/yaoxiao9/ourmp1/ServerList/servers_adr.txt"
	servers_info := get_server_info(servers_adr)
	//server_cnt := len(servers_info)
	var res [10]Grepres
	var wg sync.WaitGroup
	//wg.Add(len(servers_info))
	var log_file_name string
	for _, server := range servers_info {
		wg.Add(1)
		if test {
			log_file_name = "UnitTest" + strconv.Itoa(server.ServerId) + ".log"
		} else {
			log_file_name = "vm" + strconv.Itoa(server.ServerId) + ".log"
		}

		//fmt.Println(log_file_name)
		go connect_to_server(server, reg_pattern, log_file_name, &wg, &res)
	}
	wg.Wait()
	total_line_cnt := 0
	total_file_cnt := 0
	for _, match_res := range res {
		if match_res.Connected {
			if match_res.MatchOrnot {
				total_file_cnt += 1
				total_line_cnt += match_res.MatchCount
				//fmt.Println(match_res.GrepString)
				fmt.Println("File name:", match_res.LogfileName, "The number of matched line:", match_res.MatchCount)
			} else {
				fmt.Println("File name:", match_res.LogfileName, "The number of matched line:", match_res.MatchCount)
			}
		} else {
			fmt.Println("File name:", match_res.LogfileName, "error with connecting")
		}
	}
	fmt.Println("Total number of matched line:", total_line_cnt)
	fmt.Println("Total number of matched file:", total_file_cnt)
	return MatchRes{total_line_cnt, total_file_cnt}
}
func connect_to_server(server ServerInfo, reg_pattern string,
	log_file_name string, wg *sync.WaitGroup, res *[10]Grepres) {
	defer wg.Done()
	client, err := rpc.Dial("tcp", server.ServerAdr)
	if err != nil {
		match_res := Grepres{false, log_file_name, false, 0}
		res[server.ServerId-1] = match_res
		//wg.Done()
		log.Fatal("error with connecting to server"+strconv.Itoa(server.ServerId), err)
	}
	defer client.Close()
	request := ClientRequest{reg_pattern, log_file_name}
	var match_line_cnt string
	err = client.Call("Service.Response", request, &match_line_cnt)
	fmt.Println(match_line_cnt)
	if err != nil {
		match_res := Grepres{false, log_file_name, false, 0}
		res[server.ServerId-1] = match_res
		//wg.Done()
		log.Fatal("error with requesting to server"+strconv.Itoa(server.ServerId), err)
	}
	line_cnt, _ := strconv.Atoi(match_line_cnt)
	if line_cnt == 0 {
		match_res := Grepres{true, log_file_name, false, 0}
		res[server.ServerId-1] = match_res
	} else {
		//fmt.Println(line_cnt)
		match_res := Grepres{true, log_file_name, true, line_cnt}
		res[server.ServerId-1] = match_res
	}
	//wg.Done()
	client.Close()
}
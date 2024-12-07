package main

import (
	"akile_monitor/client/model"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

func newServer() *nbhttp.Engine {
	mux := &http.ServeMux{}
	mux.HandleFunc(cfg.WebUri, ws)
	mux.HandleFunc(cfg.UpdateUri, monitor)
	mux.HandleFunc("/delete", delete)
	mux.HandleFunc("/info", info)
	engine := nbhttp.NewEngine(nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{cfg.Listen},
		MaxLoad:                 100000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
	})
	return engine
}

func info(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getInfo(w, r)
	case "POST":
	}
}

func getInfo(w http.ResponseWriter, _ *http.Request) {
	var ret []*Host
	err := filedb.Model(&Host{}).Find(&ret).Error
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}
	by, _ := json.Marshal(&ret)
	log.Printf("get info: %s\n", by)
	w.WriteHeader(http.StatusOK)
	w.Write(by)
}

type UpdateRequest struct {
	AuthSecret string `json:"auth_secret"`
	Host
}

func updateInfo(w http.ResponseWriter, r *http.Request) {
	var ret UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&ret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "bad request")
		return
	}

	if ret.AuthSecret != cfg.AuthSecret {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "auth failed")
		return
	}

	var h Host

	filedb.Model(&Host{}).Where("name = ?", ret.Name).First(&h)
	if h.Name == "" {
		h = ret.Host
		filedb.Model(&Host{}).Create(&h)
	} else {
		h = ret.Host
		filedb.Model(&Host{}).Where("name = ?", ret.Name).Save(&h)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

type DeleteHostRequest struct {
	AuthSecret string `json:"auth_secret"`
	Name       string `json:"name"`
}

func delete(w http.ResponseWriter, r *http.Request) {
	var req DeleteHostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "bad request")
		return
	}

	if req.AuthSecret != cfg.AuthSecret {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "auth failed")
		return
	}

	var data Data
	db.Model(&Data{}).Where("name = ?", req.Name).First(&data)
	if data.Name == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "not found")
		return
	}

	db.Delete(&Data{}, "name = ?", req.Name)
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

func ws(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.NewUpgrader()
	upgrader.KeepaliveTime = time.Duration(time.Minute * 10)
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	upgrader.EnableCompression(true)
	upgrader.OnClose(close)

	upgrader.OnMessage(func(c *websocket.Conn, mt websocket.MessageType, b []byte) {
		data := fetchData()
		if err := c.WriteMessage(mt, append([]byte("data: "), data...)); err != nil {
			log.Printf("client: %s, write :%s\n", c.RemoteAddr().String(), err.Error())
			c.Close()
		}
	})

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade: %s\n", conn.RemoteAddr().String())
	}
}

func monitor(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.NewUpgrader()
	upgrader.KeepaliveTime = time.Duration(time.Second * 10)
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	upgrader.EnableCompression(true)
	upgrader.OnClose(close)
	upgrader.OnMessage(func(c *websocket.Conn, mt websocket.MessageType, b []byte) {
		if authed, ok := c.Session().(bool); ok && authed {
			var buf bytes.Buffer
			buf.Write(b)
			r, _ := gzip.NewReader(&buf)
			message, _ := io.ReadAll(r)
			r.Close()

			var d model.Data
			err := json.Unmarshal(message, &d)
			if err != nil {
				log.Printf("client: %s,unmarshal: %s\n", c.RemoteAddr().String(), err.Error())
				c.Close()
				return
			}

			var data Data
			db.Model(&Data{}).Where("name = ?", d.Host.Name).First(&data)
			if data.Name == "" {
				db.Create(&Data{Name: d.Host.Name, Data: string(message)})
			} else {
				db.Model(&Data{}).Where("name = ?", d.Host.Name).Update("data", string(message))
			}
		} else {
			if string(b) != cfg.AuthSecret {
				log.Printf("client: %s, auth failed\n", c.Conn.RemoteAddr().String())
				c.Close()
				return
			}
			c.SetSession(true)
			if err := c.WriteMessage(mt, []byte("auth success")); err != nil {
				log.Printf("client: %s, write: %s\n", c.Conn.RemoteAddr().String(), err.Error())
				c.Close()
				return
			}
		}
	})
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade: %s\n", conn.RemoteAddr().String())
	}
}

func close(c *websocket.Conn, err error) {
	log.Printf("client: %s, closed: %v\n", c.RemoteAddr().String(), err)
}

func fetchData() []byte {
	// 模拟数据获取
	var ret []Data
	db.Model(&Data{}).Find(&ret)

	var mm []model.Data

	//排序根据Name 10在9后面
	sort.Slice(ret, func(i, j int) bool {
		return compareStrings(ret[i].Name, ret[j].Name) < 0
	})

	//var jsonData string
	for _, v := range ret {
		var m model.Data
		json.Unmarshal([]byte(v.Data), &m)
		mm = append(mm, m)
	}

	jsonData, _ := json.Marshal(mm)
	return jsonData
}

// 定义一个函数来比较两个带字母和数字的字符串
func compareStrings(str1, str2 string) int {
	//先去除空格
	str1 = regexp.MustCompile(`\s+`).ReplaceAllString(str1, "")
	str2 = regexp.MustCompile(`\s+`).ReplaceAllString(str2, "")

	// 使用正则表达式提取字母和数字部分
	re := regexp.MustCompile(`([a-zA-Z]+)(\d*)`)
	matches1 := re.FindStringSubmatch(str1)
	matches2 := re.FindStringSubmatch(str2)

	if len(matches1) != 3 || len(matches2) != 3 {
		return 0 // 格式不匹配
	}

	// 提取字母部分
	letter1 := matches1[1]
	letter2 := matches2[1]

	// 提取并转换数字部分
	num1 := 0
	num2 := 0
	if len(matches1[2]) > 0 {
		num1, _ = strconv.Atoi(matches1[2])
	}
	if len(matches2[2]) > 0 {
		num2, _ = strconv.Atoi(matches2[2])
	}

	// 先比较字母部分，逐个字符比较
	for i := 0; i < len(letter1) && i < len(letter2); i++ {
		if letter1[i] < letter2[i] {
			return -1
		} else if letter1[i] > letter2[i] {
			return 1
		}
	}

	// 如果字母部分相同，长度不等时，短的字母部分小
	if len(letter1) < len(letter2) {
		return -1
	} else if len(letter1) > len(letter2) {
		return 1
	}

	// 字母相同，比较数字部分
	if num1 < num2 {
		return -1
	} else if num1 > num2 {
		return 1
	}

	// 如果字母和数字都相同
	return 0
}

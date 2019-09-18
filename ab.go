package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var (
	server           *http.Server
	listener         net.Listener
	graceful         = flag.Bool("graceful", false, "graceful restart")
	config_file      = flag.String("c", "./config.json", "use config file")
	conf             *Config
	mylogger         *ZdLogger
	paramNameVersion = "__abv"
	paramNameData    = "__abd"
	sockFile         = "/tmp/abtest.sock"
	uuid             *ZdUUID
	bufferSize       = 1024 * 32
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, bufferSize)
	},
}

func main() {
	flag.Parse()
	conf = NewConfig(*config_file).Parse()
	mylogger = NewLogger(conf.GetLogDir(), conf.GetLogFormat(), conf.GetLogPrefix())
	if conf.Default != nil {
		if v, ok := conf.Default["paramNameVersion"]; ok {
			paramNameVersion = v.(string)
		}
		if v, ok := conf.Default["paramNameData"]; ok {
			paramNameData = v.(string)
		}
		if v, ok := conf.Default["sockFile"]; ok {
			sockFile = v.(string)
		}
	}
	uuid = NewUUID()
	start()
}

func handleSignal() {
	abSignal := make(chan os.Signal, 1)
	signal.Notify(abSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		sig := <-abSignal
		log.Printf("signal receive: %v\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM: // 终止进程执行
			log.Println("shutdown")
			signal.Stop(abSignal)
			server.SetKeepAlivesEnabled(false)
			if err := server.Shutdown(ctx); err != nil {
				mylogger.Println(err)
			}
			log.Println("graceful shutdown")
			return
		case syscall.SIGUSR1: //重新加载配置文件
			log.Println("reload config file")
			conf.Parse()
			mylogger.Println(conf)
			continue
		case syscall.SIGUSR2: // 进程热重启
			log.Println("reload")
			err := reload()
			if err != nil {
				log.Fatalf("graceful reload error: %v", err)
			}
			if err := server.Shutdown(ctx); err != nil {
				mylogger.Println(err)
			}
			log.Println("graceful reload")
			return
		}
	}
}

func reload() error {
	tl, ok := listener.(*net.TCPListener)
	if !ok {
		return errors.New("listener is not tcp listener")
	}
	f, err := tl.File()
	if err != nil {
		return err
	}

	args := []string{"-graceful", "-c=" + *config_file}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{f}
	return cmd.Start()
}

func start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/abtest_config_reload", func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				writer.Write([]byte("reload fail"))
				mylogger.Println(r)
			}
		}()
		conf.Parse()
		writer.Write([]byte("reload success"))
	})
	mux.HandleFunc("/slb_check", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("ok"))
	})
	mux.HandleFunc("/", proxy)

	defer func() {
		if r := recover(); r != nil {
			mylogger.Println(r)
		}
	}()

	var port int
	if v, ok := conf.Default["port"]; ok {
		port = int(v.(float64))
	}

	server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 60,
		IdleTimeout:  time.Second * 300,
		Handler:      mux,
		// Handler: http.TimeoutHandler(mux, time.Second*60, "TimeOut"),
	}

	var err error
	if *graceful {
		f := os.NewFile(3, "")
		listener, err = net.FileListener(f)
	} else {
		listener, err = net.Listen("tcp", server.Addr)
	}
	if err != nil {
		mylogger.Fatalf("listener error: %v\n", err)
	}

	go func() {
		err = server.Serve(listener)
		if err != nil {
			// 正常退出
			if err == http.ErrServerClosed {
				log.Printf("Server closed under request(%s)\n", err)
			} else {
				log.Printf("Server closed unexpected(%s)\n", err)
			}
		}
	}()

	ioutil.WriteFile(sockFile, []byte(strconv.Itoa(os.Getpid())), os.ModeAppend)

	log.Printf("Starting httpServer pid:%d, port:%d\n", os.Getpid(), port)
	go func() {
		log.Println(http.ListenAndServe(":10000", nil))
	}()
	handleSignal()
	log.Println("Server exited")
}

func proxy(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			mylogger.Println(r)
			mylogger.Println(string(debug.Stack()))
		}
	}()

	var __abv, __abd string
	if tmp, ok := r.Header[paramNameVersion]; ok {
		__abv = tmp[0]
	}
	if __abv == "" {
		if tmp, err := r.Cookie(paramNameVersion); err == nil {
			__abv = tmp.Value
		}
	}
	if tmp, ok := r.Header[paramNameData]; ok {
		__abd = tmp[0]
	}
	if __abd == "" {
		if tmp, err := r.Cookie(paramNameData); err == nil {
			__abd = tmp.Value
		}
	}

	ip := __getIp(r.Host, __abv, __abd)
	tmp_url := "http://" + ip + r.URL.String()

	tmp_uuid := uuid.createUUID()
	r.Header.Add("AB-REQUEST-ID", tmp_uuid)
	writeLog(r, tmp_url)

	req, err := http.NewRequest(r.Method, tmp_url, r.Body)

	if err != nil {
		errStr := tmp_uuid + " backend server error1"
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(errStr))
		ret, _ := json.Marshal(r.Header)
		mylogger.Println(ret, errStr)
		return
	}

	req.Host = r.Host
	req.URL.Scheme = "http"
	for _, v := range r.Cookies() {
		req.AddCookie(v)
	}

	req.Header = r.Header
	req.Header.Add("AB-REQUEST-ID", tmp_uuid)
	client := &http.Client{
		Timeout: time.Second * 60,
	}
	resp, err := client.Do(req)

	if err != nil {
		errStr := tmp_uuid + " backend server error2"
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(errStr))
		mylogger.Println(err, errStr)
		return
	}

	for k, v := range resp.Header {
		for _, value := range v {
			w.Header().Add(k, value)
		}
	}
	for _, cookie := range resp.Cookies() {
		w.Header().Add("Set-Cookie", cookie.Raw)
	}
	w.Header().Add("AB-REQUEST-ID", tmp_uuid)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	// buffer := getBuffer()
	// defer putBuffer(buffer)
	// io.CopyBuffer(w, resp.Body, buffer)
	resp.Body.Close()
	r.Body.Close()
}

func writeLog(r *http.Request, url string) {
	reqBytes, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBytes))
	ret, _ := json.Marshal(r.Header)

	ct := r.Header.Get("Content-Type")
	ct, _, _ = mime.ParseMediaType(ct)

	logout := make([]interface{}, 0, 10)
	logout = append(logout, r.Method, r.Host, r.URL, url, fmt.Sprintf("LOG_HEADER: %s", ret))

	switch ct {
	case "application/x-www-form-urlencoded":
		r.ParseForm()
		ret1, _ := json.Marshal(r.Form)
		ret2, _ := json.Marshal(r.PostForm)
		logout = append(logout, fmt.Sprintf("LOG_FORM: %s", ret1), fmt.Sprintf("LOG_POSTFORM: %s", ret2))
	case "multipart/form-data":
		r.ParseMultipartForm(1 << 16)
		ret3, _ := json.Marshal(r.MultipartForm)
		logout = append(logout, fmt.Sprintf("LOG_MULTIPARTFORM: %s", ret3))
	case "application/json":
		logout = append(logout, fmt.Sprintf("LOG_JSON: %s", reqBytes))
	default:
		logout = append(logout, fmt.Sprintf("LOG_OTHER: %s", reqBytes))
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBytes))

	mylogger.Println(logout...) //记录请求数据

}

func __abdDecode(host, data string) []int64 {
	// data := "4a337757333d33232445333d337362704863333d33727c7f672064333d33746252475f74334c"

	var secrets = conf.GetDefaultSecret()

	if v, ok := conf.Rule[host]; ok {
		if v.Secret != nil {
			secrets = v.Secret
		}
	}
	if secrets == nil {
		return nil
	}
	c, _ := hex.DecodeString(data)

	l := len(secrets)
	field_num := 4
	ret := make(chan []string, l)

	wg := sync.WaitGroup{}

	for i := 0; i < l; i++ {
		wg.Add(1)
		go func(i int, c []byte) {
			defer wg.Done()
			c1 := AbDecode([]byte(secrets[i]), c)
			demo1 := make([]string, field_num)
			json.Unmarshal(c1, &demo1)
			if demo1[0] != "" {
				ret <- demo1
			} else {
				ret <- []string{}
			}
		}(i, c)
	}
	wg.Wait()
	close(ret)
	result := make([]int64, 0, 10)

	for v1 := range ret {
		if len(v1) > 0 {
			for _, v2 := range v1 {
				ret := anyToDecimal(v2, 64)
				result = append(result, int64(ret))
			}
			break
		}
	}
	return result
}

//综合所有条件，得到反向代理目标服务器的ip
func __getIp(host, abv, __abd string) string {
	IP_defaultA := conf.GetDefaultARandIp()

	//所有配置都没有
	var hostParams *ConfigRuleOK
	var ok bool
	if hostParams, ok = conf.RuleOK[host]; !ok {
		return IP_defaultA
	}

	IP_defaultB := conf.GetDefaultBRandIp()
	if v, ok := conf.Rule[host]; ok {
		if v.GroupA != nil {
			IP_defaultA = conf.GetGroupARandIp(host)
		}
		if v.GroupB != nil {
			IP_defaultB = conf.GetGroupBRandIp(host)
		}
	}

	//所有标识都没有
	if abv == "" && __abd == "" {
		return IP_defaultA
	}

	var abd []int64
	//解密标识信息
	if __abd != "" {
		abd = __abdDecode(host, __abd)
	}
	abd_len := len(abd)

	if hostParams.Version.Has(abv) && abd_len == 0 { //只有版本号
		return IP_defaultB
	}

	if hostParams.Version.Has(abv) || conf.GetVersions(host) == nil { //命中版本号，或根本没配置版本号
		if abd_len > 0 && int64(abd[0]) < time.Now().Unix() { //过期失效
			return IP_defaultA
		}
		if abd_len > 1 && hostParams.Uid.Has(abd[1]) { //用uid判断
			return IP_defaultB
		} else if abd_len > 2 && hostParams.Telphone.Has(abd[2]) { //用telphone判断
			return IP_defaultB
		} else if abd_len > 3 && hostParams.City.Has(abd[3]) { //用city判断
			return IP_defaultB
		}
	}

	return IP_defaultA
}

func getBuffer() []byte {
	if bf, ok := bufferPool.Get().([]byte); ok {
		return bf
	}
	return make([]byte, bufferSize)
}

func putBuffer(bt []byte) {
	bufferPool.Put(bt)
}

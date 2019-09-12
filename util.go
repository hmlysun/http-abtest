package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func AbEncode(secret_key_1, sk []byte) []byte {
	ret := make([]byte, len(sk))
	for k1, v1 := range sk {
		tmp := v1
		for _, v2 := range secret_key_1 {
			tmp ^= v2
		}
		ret[k1] = tmp
	}
	return ret
}

func AbDecode(secret_key_1, sk []byte) []byte {
	return AbEncode(secret_key_1, sk)
}

//var tenToAny map[int]string = map[int]string{0: "0", 1: "1", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7", 8: "8", 9: "9", 10: "a", 11: "b", 12: "c", 13: "d", 14: "e", 15: "f", 16: "g", 17: "h", 18: "i", 19: "j", 20: "k", 21: "l", 22: "m", 23: "n", 24: "o", 25: "p", 26: "q", 27: "r", 28: "s", 29: "t", 30: "u", 31: "v", 32: "w", 33: "x", 34: "y", 35: "z", 36: ":", 37: ";", 38: "<", 39: "=", 40: ">", 41: "?", 42: "@", 43: "[", 44: "]", 45: "^", 46: "_", 47: "{", 48: "|", 49: "}", 50: "A", 51: "B", 52: "C", 53: "D", 54: "E", 55: "F", 56: "G", 57: "H", 58: "I", 59: "J", 60: "K", 61: "L", 62: "M", 63: "N", 64: "O", 65: "P", 66: "Q", 67: "R", 68: "S", 69: "T", 70: "U", 71: "V", 72: "W", 73: "X", 74: "Y", 75: "Z"}

var tenToAny map[int]string = map[int]string{
	0: "0", 1: "1", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7", 8: "8", 9: "9",
	10: "a", 11: "b", 12: "c", 13: "d", 14: "e", 15: "f", 16: "g", 17: "h", 18: "i",
	19: "j", 20: "k", 21: "l", 22: "m", 23: "n", 24: "o", 25: "p", 26: "q", 27: "r",
	28: "s", 29: "t", 30: "u", 31: "v", 32: "w", 33: "x", 34: "y", 35: "z", 36: "A",
	37: "B", 38: "C", 39: "D", 40: "E", 41: "F", 42: "G", 43: "H", 44: "I", 45: "J",
	46: "K", 47: "L", 48: "M", 49: "N", 50: "O", 51: "P", 52: "Q", 53: "R", 54: "S",
	55: "T", 56: "U", 57: "V", 58: "W", 59: "X", 60: "Y", 61: "Z", 62: "-", 63: "=",
}

//func main() {
//	fmt.Println(decimalToAny(9999, 76))
//	fmt.Println(anyToDecimal("1F[", 76))
//}

// 10进制转任意进制
func decimalToAny(num, n int) string {
	new_num_str := ""
	var remainder int
	var remainder_string string
	for num != 0 {
		remainder = num % n
		if 64 > remainder && remainder > 9 {
			remainder_string = tenToAny[remainder]
		} else {
			remainder_string = strconv.Itoa(remainder)
		}
		new_num_str = remainder_string + new_num_str
		num = num / n
	}
	return new_num_str
}

// map根据value找key
func findkey(in string) int {
	result := -1
	for k, v := range tenToAny {
		if in == v {
			result = k
		}
	}
	return result
}

// 任意进制转10进制
func anyToDecimal(num string, n int) int {
	var new_num float64
	new_num = 0.0
	nNum := len(strings.Split(num, "")) - 1
	for _, value := range strings.Split(num, "") {
		tmp := float64(findkey(value))
		if tmp != -1 {
			new_num = new_num + tmp*math.Pow(float64(n), float64(nNum))
			nNum = nNum - 1
		} else {
			break
		}
	}
	return int(new_num)
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.2"
}

func GetMacAddrs() (string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}
		return macAddr, nil
	}
	return "", err
}

type ZdUUID struct {
	base_hash hash.Hash
	buffer    bytes.Buffer
	mutex     sync.Mutex
	rand_part map[string]interface{} //动态随机因子
	base_part string                 //固定随机因子
}

func NewUUID() *ZdUUID {
	obj := &ZdUUID{
		base_hash: md5.New(),
		buffer:    bytes.Buffer{},
		mutex:     sync.Mutex{},
		rand_part: make(map[string]interface{}),
	}

	host, _ := os.Hostname()
	mac, _ := GetMacAddrs()
	obj.base_part = host + mac + GetLocalIP()
	obj.rand_part["time"] = func() interface{} {
		return time.Now().UnixNano()
	}
	obj.rand_part["rand"] = func() interface{} {
		rand.Seed(time.Now().UnixNano())
		return rand.Intn(1<<63 - 1)
	}
	return obj
}

func (this *ZdUUID) reset() {
	this.base_part = ""
	this.rand_part = make(map[string]interface{})
}

//可以从外部重新自定义随机因子
func (this *ZdUUID) setConfig(base_part string, rand_part map[string]interface{}) {
	this.reset()
	this.base_part = base_part
	this.rand_part = rand_part
}

//go tool pprof -http=:8000 http://localhost:10000/debug/pprof/heap
func (this *ZdUUID) createUUID() string {
	this.mutex.Lock()
	this.buffer.Reset()
	this.buffer.WriteString(this.base_part)
	for _, v := range this.rand_part {
		this.buffer.WriteString(fmt.Sprintf("%v", v.(func() interface{})()))
	}
	this.base_hash.Reset()
	io.WriteString(this.base_hash, this.buffer.String())
	hashstr := fmt.Sprintf("%X", this.base_hash.Sum(nil))
	tmp := make([]string, 0, 7)
	for i := 0; i < 4; i++ {
		tmp = append(tmp, hashstr[i*8:i*8+8])
	}
	this.mutex.Unlock()
	return strings.Join(tmp, "-")
	// this.buffer.Reset()
	// this.buffer.WriteString(hashstr[0:8])
	// this.buffer.WriteByte('-')
	// this.buffer.WriteString(hashstr[8:16])
	// this.buffer.WriteByte('-')
	// this.buffer.WriteString(hashstr[16:24])
	// this.buffer.WriteByte('-')
	// this.buffer.WriteString(hashstr[24:32])
	// this.mutex.Unlock()
	// return this.buffer.String()
}

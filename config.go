package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"
)

type Config struct {
	FilePath      string
	Log           map[string]string      `json:"log"`
	Default       map[string]interface{} `json:"defaultOption"`
	DefaultServer map[string][]string    `json:"defaultServer"`
	DefaultSecret []string               `json:"defaultSecret"`
	Rule          map[string]ConfigRule  `json:"rule"`
	RuleOK        map[string]*ConfigRuleOK
}

type ConfigRule struct {
	Secret   []string `json:"secrets"`
	Version  []string `json:"versions"`
	Uid      []int64  `json:"uids"`
	Telphone []int64  `json:"telphones"`
	City     []int64  `json:"citys"`
	Field1   []int64  `json:"field1"`
	Field2   []int64  `json:"field2"`
	Field3   []int64  `json:"field3"`
	GroupA   []string `json:"groupA"`
	GroupB   []string `json:"groupB"`
}

type ConfigRuleOK struct {
	Version  *SetMap
	Uid      *SetMap
	Telphone *SetMap
	City     *SetMap
	Field1   *SetMap
	Field2   *SetMap
	Field3   *SetMap
}

func NewConfig(filePath string) *Config {
	return &Config{
		FilePath: filePath,
	}
}

func (this *Config) Parse() *Config {
	f, err := os.Open(this.FilePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(this); err != nil {
		log.Fatalln(err)
	}
	if this.RuleOK == nil {
		this.RuleOK = make(map[string]*ConfigRuleOK)
	}
	if this.Rule != nil {
		for k, v := range this.Rule {
			tmp := &ConfigRuleOK{}
			if v.Version != nil {
				tmp_1 := NewSet()
				for _, v1 := range v.Version {
					tmp_1.Add(v1)
				}
				tmp.Version = tmp_1
			}
			if v.Uid != nil {
				tmp_1 := NewSet()
				for _, v1 := range v.Uid {
					tmp_1.Add(v1)
				}
				tmp.Uid = tmp_1
			}
			if v.Telphone != nil {
				tmp_1 := NewSet()
				for _, v1 := range v.Telphone {
					tmp_1.Add(v1)
				}
				tmp.Telphone = tmp_1
			}
			if v.City != nil {
				tmp_1 := NewSet()
				for _, v1 := range v.City {
					tmp_1.Add(v1)
				}
				tmp.City = tmp_1
			}
			if v.Field1 != nil {
				tmp_1 := NewSet()
				for _, v1 := range v.Field1 {
					tmp_1.Add(v1)
				}
				tmp.Field1 = tmp_1
			}
			if v.Field2 != nil {
				tmp_1 := NewSet()
				for _, v1 := range v.Field2 {
					tmp_1.Add(v1)
				}
				tmp.Field2 = tmp_1
			}
			if v.Field3 != nil {
				tmp_1 := NewSet()
				for _, v1 := range v.Field3 {
					tmp_1.Add(v1)
				}
				tmp.Field3 = tmp_1
			}
			this.RuleOK[k] = tmp
		}
	}
	return this
}

func (this *Config) GetLogDir() (ret string) {
	if this.Log != nil {
		if v, ok := this.Log["dir"]; ok {
			ret = v
		}
	}
	return
}

func (this *Config) GetLogFormat() (ret string) {
	if this.Log != nil {
		if v, ok := this.Log["format"]; ok {
			ret = v
		}
	}
	return
}

func (this *Config) GetLogPrefix() (ret string) {
	if this.Log != nil {
		if v, ok := this.Log["prefix"]; ok {
			ret = v
		}
	}
	return
}

func (this *Config) GetDefaultServer() (ret map[string][]string) {
	if this.DefaultServer != nil {
		ret = this.DefaultServer
	}
	return
}

func (this *Config) GetDefaultServerGroupA() (ret []string) {
	all := this.GetDefaultServer()
	if all != nil {
		if v, ok := all["groupA"]; ok {
			ret = v
		}
	}
	return
}

func (this *Config) GetDefaultServerGroupB() (ret []string) {
	all := this.GetDefaultServer()
	if all != nil {
		if v, ok := all["groupB"]; ok {
			ret = v
		}
	}
	return
}

func (this *Config) GetDefaultSecret() (ret []string) {
	if this.DefaultSecret != nil {
		return this.DefaultSecret
	}
	return
}

func (this *Config) GetVersions(host string) (ret []string) {
	if this.Rule != nil {
		if v, ok := this.Rule[host]; ok {
			ret = v.Version
		}
	}
	return
}

func (this *Config) GetUids(host string) (ret []int64) {
	if this.Rule != nil {
		if v, ok := this.Rule[host]; ok {
			ret = v.Uid
		}
	}
	return
}

func (this *Config) GetTelphones(host string) (ret []int64) {
	if this.Rule != nil {
		if v, ok := this.Rule[host]; ok {
			ret = v.Telphone
		}
	}
	return
}

func (this *Config) GetCitys(host string) (ret []int64) {
	if this.Rule != nil {
		if v, ok := this.Rule[host]; ok {
			ret = v.City
		}
	}
	return
}

func (this *Config) GetSecrets(host string) (ret []string) {
	if this.Rule != nil {
		if v, ok := this.Rule[host]; ok {
			ret = v.Secret
		}
	}
	return
}

func (this *Config) GetGroupA(host string) (ret []string) {
	if this.Rule != nil {
		if v, ok := this.Rule[host]; ok {
			ret = v.GroupA
		}
	}
	return
}

func (this *Config) GetGroupB(host string) (ret []string) {
	if this.Rule != nil {
		if v, ok := this.Rule[host]; ok {
			ret = v.GroupB
		}
	}
	return
}

func (this *Config) GetDefaultARandIp() (ret string) {
	ips := this.GetDefaultServerGroupA()
	if ips != nil {
		groupA_len := len(ips)
		if groupA_len > 0 {
			rand.Seed(time.Now().UnixNano())
			ret = ips[rand.Intn(groupA_len)]

		}
	}
	return
}

func (this *Config) GetDefaultBRandIp() (ret string) {
	ips := this.GetDefaultServerGroupB()
	if ips != nil {
		groupB_len := len(ips)
		if groupB_len > 0 {
			rand.Seed(time.Now().UnixNano())
			ret = ips[rand.Intn(groupB_len)]

		}
	}
	return
}

func (this *Config) GetGroupARandIp(host string) (ret string) {
	ips := this.GetGroupA(host)
	if ips != nil {
		groupA_len := len(ips)
		if groupA_len > 0 {
			rand.Seed(time.Now().UnixNano())
			ret = ips[rand.Intn(groupA_len)]

		}
	}
	return
}

func (this *Config) GetGroupBRandIp(host string) (ret string) {
	ips := this.GetGroupB(host)
	if ips != nil {
		groupB_len := len(ips)
		if groupB_len > 0 {
			rand.Seed(time.Now().UnixNano())
			ret = ips[rand.Intn(groupB_len)]

		}
	}
	return
}

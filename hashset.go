package main

import (
	"sync"
)

type ZeroValue struct{}

type SetMap struct {
	*sync.RWMutex
	items map[interface{}]ZeroValue
}

func NewSet() *SetMap {
	return &SetMap{
		&sync.RWMutex{},
		make(map[interface{}]ZeroValue),
	}
}

//添加
func (this *SetMap) Add(v ...interface{}) (ret bool) {
	this.Lock()
	defer this.Unlock()
	ret = true
	for _, _v := range v {
		this.items[_v] = ZeroValue{}
	}
	return
}

//查询
func (this *SetMap) Has(v interface{}) (ret bool) {
	ret = false
	if _, ok := this.items[v]; ok {
		ret = true
	}
	return
}

//删除
func (this *SetMap) Remove(v ...interface{}) (ret bool) {
	this.Lock()
	defer this.Unlock()
	ret = true
	for _, _v := range v {
		delete(this.items, _v)
	}
	return
}

//清空重置
func (this *SetMap) Clear() (ret bool) {
	this.Lock()
	defer this.Unlock()
	ret = true
	this.items = make(map[interface{}]ZeroValue)
	return
}

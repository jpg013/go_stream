package main

import (
	"container/list"
	"sync"

	"github.com/jpg013/go_stream/types"
)

type Buffer interface {
	Write(types.Chunk) int
	Read() (types.Chunk, int)
	Len() int
}

type BufferType struct {
	list *list.List
	mux  sync.RWMutex
}

func (b *BufferType) Write(item types.Chunk) int {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.list.PushBack(item)
	return b.list.Len()
}

func (b *BufferType) Read() (types.Chunk, int) {
	b.mux.Lock()
	defer b.mux.Unlock()
	item := b.list.Front()
	if item == nil || item.Value == nil {
		return item, 0
	}
	b.list.Remove(item)
	return item.Value, b.list.Len()
}

func (b *BufferType) Len() int {
	b.mux.RLock()
	len := b.list.Len()
	defer b.mux.RUnlock()
	return len
}

func NewBuffer() Buffer {
	list := list.New()

	return &BufferType{
		list: list,
	}
}

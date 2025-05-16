package books

import (
	"container/list"
	"errors"
	"sync"
)

type Book struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Pages uint32 `json:"pages"`
}

type Library interface {
	Get(string) (*Book, error)
	Add(*Book) error
	Remove(string) error
}

type library struct {
	list  *list.List
	index map[string]*list.Element
	mu    sync.RWMutex // Added for thread safety
}

func NewLibrary() Library {
	return &library{
		list:  list.New(),
		index: make(map[string]*list.Element),
	}
}

func (l *library) Get(id string) (*Book, error) {
	l.mu.RLock()         // Read lock
	defer l.mu.RUnlock() // Unlock when done

	book, ok := l.index[id]
	if !ok {
		return nil, errors.New("book not found")
	}

	return book.Value.(*Book), nil
}

func (l *library) Add(book *Book) error {
	l.mu.Lock()         // Write lock
	defer l.mu.Unlock() // Unlock when done

	_, ok := l.index[book.Id]
	if ok {
		return errors.New("book already exists")
	}

	l.index[book.Id] = l.list.PushBack(book)
	return nil
}

func (l *library) Remove(id string) error {
	l.mu.Lock()         // Write lock
	defer l.mu.Unlock() // Unlock when done

	book, ok := l.index[id]
	if !ok {
		return errors.New("book not found")
	}

	l.list.Remove(book)
	delete(l.index, id)

	return nil
}

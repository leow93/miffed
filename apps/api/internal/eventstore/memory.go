package eventstore

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Event struct {
	eventType string
	data      []byte
}

type Category struct {
	position uint64
	events   *[]Event
	mutex    *sync.Mutex
	streams  map[string]*Stream
}

func NewCategory(lock *sync.Mutex) *Category {
	var events []Event
	return &Category{
		events:   &events,
		position: 0,
		mutex:    lock,
		streams:  make(map[string]*Stream),
	}
}

func (cat *Category) sync(streamName string, expectedVersion uint64, events []Event) error {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()
	s, ok := cat.streams[streamName]
	if !ok {
		// create the stream
		s = NewStream()
		cat.streams[streamName] = s
	}

	if expectedVersion != s.position {
		return fmt.Errorf("wrong expected version: got %d, want %d", s.position, expectedVersion)
	}
	// append to stream
	newStreamEvents := append(*s.events, events...)
	s.events = &newStreamEvents
	s.position += uint64(len(events))
	// ...then append to category
	newCategoryEvents := append(*cat.events, events...)
	cat.events = &newCategoryEvents
	cat.position += uint64(len(events))

	return nil
}

type Stream struct {
	events   *[]Event
	position uint64
}

func NewStream() *Stream {
	var events []Event
	return &Stream{
		events:   &events,
		position: 0,
	}
}

func (s *Stream) read() []Event {
	return *s.events
}

type MemoryStore struct {
	categories    map[string]*Category
	categoryLocks map[string]*sync.Mutex
	globalLock    sync.Mutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		categories:    make(map[string]*Category),
		categoryLocks: make(map[string]*sync.Mutex),
		globalLock:    sync.Mutex{},
	}
}

func (store *MemoryStore) getCategoryLock(categoryName string) *sync.Mutex {
	// Use global lock to acquire stream lock from map
	store.globalLock.Lock()
	defer store.globalLock.Unlock()
	lock, ok := store.categoryLocks[categoryName]

	if !ok {
		lock = &sync.Mutex{}
		store.categoryLocks[categoryName] = lock
	}
	return lock
}

func parseCategory(streamName string) (string, error) {
	parts := strings.Split(streamName, "-")
	if len(parts) < 2 {
		return "", errors.New("no category")
	}
	category := parts[0]
	if category == "" {
		return "", errors.New("empty category")
	}
	return category, nil
}

func (store *MemoryStore) getCategoryByStreamName(streamName string) (*Category, error) {
	categoryName, err := parseCategory(streamName)
	if err != nil {
		return nil, err
	}
	lock := store.getCategoryLock(categoryName)

	store.globalLock.Lock()
	category, ok := store.categories[categoryName]
	if !ok {
		category = NewCategory(lock)
		store.categories[categoryName] = category
	}
	store.globalLock.Unlock()

	return category, nil
}

func (store *MemoryStore) getCategory(categoryName string) (*Category, error) {
	store.globalLock.Lock()
	defer store.globalLock.Unlock()

	category, ok := store.categories[categoryName]
	if !ok {
		return nil, errors.New("no such category")
	}
	return category, nil
}

func (store *MemoryStore) getStream(streamName string) (*Stream, error) {
	category, err := store.getCategoryByStreamName(streamName)
	if err != nil {
		return nil, err
	}
	// Lock the category so we can safely read from the stream map
	category.mutex.Lock()
	defer category.mutex.Unlock()
	stream, ok := category.streams[streamName]
	if !ok {
		stream = NewStream()
		category.streams[streamName] = stream
	}

	return stream, nil
}

func (store *MemoryStore) AppendToStream(streamName string, expectedVersion uint64, events []Event) error {
	cat, err := store.getCategoryByStreamName(streamName)
	if err != nil {
		return err
	}
	return cat.sync(streamName, expectedVersion, events)
}

func (store *MemoryStore) ReadStream(streamName string) ([]Event, uint64, error) {
	stream, err := store.getStream(streamName)
	if err != nil {
		return []Event{}, 0, err
	}
	evs := stream.read()
	return evs, stream.position, nil
}

func (store *MemoryStore) ReadCategory(categoryName string, fromPosition uint64) ([]Event, error) {
	cat, err := store.getCategory(categoryName)
	if err != nil {
		return []Event{}, err
	}

	evs := *cat.events
	evs = evs[fromPosition:]

	return evs, nil
}

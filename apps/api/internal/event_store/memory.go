package eventstore

import "sync"

type Message struct {
	Type string
	Data []byte
}

type Stream struct {
	id       string
	messages []Message
	mutex    sync.Mutex
}

func (s *Stream) Sync(messages []Message, expectedVersion int64) (AppendResult, error) {
	s.messages = append(s.messages, messages...)
	return AppendResult{Position: int64(len(s.messages))}, nil
}

type MemoryStore struct {
	streams map[string]*Stream
}

func NewStore() *MemoryStore {
	return &MemoryStore{
		streams: make(map[string]*Stream),
	}
}

type AppendResult struct {
	Position int64
}

func (store *MemoryStore) loadStream(streamName string) *Stream {
	if s, ok := store.streams[streamName]; ok {
		return s
	}

	stream := &Stream{
		id:       streamName,
		messages: []Message{},
		mutex:    sync.Mutex{},
	}
	store.streams[streamName] = stream
	return stream
}

func (store *MemoryStore) AppendToStream(stream string, messages []Message, expectedVersion int64) (AppendResult, error) {
	s := store.loadStream(stream)
	return s.Sync(messages, expectedVersion)
}

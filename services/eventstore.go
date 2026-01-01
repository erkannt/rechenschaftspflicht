package services

import "sync"

type Event struct {
	Tag       string `json:"tag"`
	Comment   string `json:"comment"`
	Value     string `json:"value"`
	CreatedAt string `json:"createdAt"`
}

type EventStore interface {
	Record(event Event) error
	GetAll() ([]Event, error)
}

type InMemoryEventStore struct {
	mu     sync.Mutex
	events []Event
}

func NewEventStore() EventStore {
	return &InMemoryEventStore{
		events: make([]Event, 0),
	}
}

func (s *InMemoryEventStore) Record(event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

func (s *InMemoryEventStore) GetAll() ([]Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make([]Event, len(s.events))
	copy(copied, s.events)
	return copied, nil
}

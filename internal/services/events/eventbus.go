package events

import (
	"sync"

	"github.com/google/uuid"
)

type EventType string

const (
	PlaylistImportStarted   EventType = "playlist_import_started"
	PlaylistImportCompleted EventType = "playlist_import_completed"
	PlaylistImportFailed    EventType = "playlist_import_failed"
)

type Event struct {
	Type   EventType
	UserID uuid.UUID
	Data   interface{}
}

type EventBus struct {
	subscribers map[EventType][]chan Event // For each event type, keep track of all subscriber channels that should receive those events.
	mu          sync.RWMutex
}

var globalEventBus *EventBus

func init() {
	globalEventBus = &EventBus{
		subscribers: make(map[EventType][]chan Event),
	}
}

func NewEventBus() *EventBus {
	return globalEventBus
}

func GetGlobalEventBus() *EventBus {
	return globalEventBus
}

func (eb *EventBus) Subscribe(eventType EventType, bufferSize int) <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, bufferSize)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// check if subscriber to the event to publish
	if subscribers, exists := eb.subscribers[event.Type]; exists {
		// for each subscriber, send event to its channel
		for _, ch := range subscribers {
			select {
			case ch <- event:
			default:
				// Channel is full, skip this subscriber to avoid blocking
			}
		}
	}
}

// UserEventSubscriber manages SSE connections for a specific user
type UserEventSubscriber struct {
	userID   uuid.UUID
	eventBus *EventBus
	channels []<-chan Event
}

func NewUserEventSubscriber(userID uuid.UUID, eventBus *EventBus) *UserEventSubscriber {
	return &UserEventSubscriber{
		userID:   userID,
		eventBus: eventBus,
		channels: make([]<-chan Event, 0),
	}
}

func (ues *UserEventSubscriber) SubscribeToPlaylistEvents() <-chan Event {
	// Create a single channel that receives all playlist-related events for this user
	eventChan := make(chan Event, 10)

	// Subscribe to all playlist events
	importStarted := ues.eventBus.Subscribe(PlaylistImportStarted, 5)
	importCompleted := ues.eventBus.Subscribe(PlaylistImportCompleted, 5)
	importFailed := ues.eventBus.Subscribe(PlaylistImportFailed, 5)

	// Start goroutine to filter events for this user and forward to the unified channel
	go func() {
		defer close(eventChan)

		for {
			select {
			case event, ok := <-importStarted:
				if !ok {
					return
				}
				if event.UserID == ues.userID {
					eventChan <- event
				}
			case event, ok := <-importCompleted:
				if !ok {
					return
				}
				if event.UserID == ues.userID {
					eventChan <- event
				}
			case event, ok := <-importFailed:
				if !ok {
					return
				}
				if event.UserID == ues.userID {
					eventChan <- event
				}
			}
		}
	}()

	return eventChan
}

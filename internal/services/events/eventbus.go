package events

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type EventType string

const (
	PlaylistImportCompleted EventType = "playlist_import_completed"
	PlaylistCreated         EventType = "playlist_created"
	PlaylistUpdated         EventType = "playlist_updated"
	PlaylistDeleted         EventType = "playlist_deleted"
	PlaylistSongAdded       EventType = "playlist_song_added"
	PlaylistSongRemoved     EventType = "playlist_song_removed"
	PlaylistSyncCompleted   EventType = "playlist_sync_completed"
)

type Event struct {
	Type   EventType
	UserID uuid.UUID
	Data   interface{}
}

type EventBus struct {
	userSubscribers map[uuid.UUID][]chan Event // For each user, keep track of all their event channels
	mu              sync.RWMutex
}

var globalEventBus *EventBus

func init() {
	globalEventBus = &EventBus{
		userSubscribers: make(map[uuid.UUID][]chan Event),
	}
}

func NewEventBus() *EventBus {
	return globalEventBus
}

func GetGlobalEventBus() *EventBus {
	return globalEventBus
}

func (eb *EventBus) SubscribeUser(userID uuid.UUID) <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, 20)
	eb.userSubscribers[userID] = append(eb.userSubscribers[userID], ch)
	return ch
}

func (eb *EventBus) UnsubscribeUser(userID uuid.UUID, ch <-chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if channels, exists := eb.userSubscribers[userID]; exists {
		// Iterate backwards to safely modify slice during iteration
		for i := len(channels) - 1; i >= 0; i-- {
			// Compare by converting both to the same type for comparison
			if (<-chan Event)(channels[i]) == ch {
				// Safe close with panic protection
				func() {
					defer func() {
						recover() // Ignore panic if channel already closed
					}()
					close(channels[i])
				}()

				// Safe slice removal
				eb.userSubscribers[userID] = append(channels[:i], channels[i+1:]...)
				break
			}
		}
		// Clean up empty user entries
		if len(eb.userSubscribers[userID]) == 0 {
			delete(eb.userSubscribers, userID)
		}
	}
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	fmt.Println("calling", event.UserID, event.Type)

	// Send event directly to all channels for this user
	if channels, exists := eb.userSubscribers[event.UserID]; exists {
		for _, ch := range channels {
			select {
			case ch <- event:
			default:
				// Channel is full, skip this subscriber to avoid blocking
			}
		}
	}
}

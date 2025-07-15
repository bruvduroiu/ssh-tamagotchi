package main

import (
	"fmt"
	"sync"
	"time"
)

type Pet struct {
	Name      string
	Health    int
	MaxHealth int
	Mood      string
	LastFed   time.Time
	LastPet   time.Time
	Actions   []Action
}

type Action struct {
	User      string
	Type      string
	Timestamp time.Time
}

type GameState struct {
	Pet            Pet
	ConnectedUsers map[string]bool
	Subscribers    map[string]chan GameUpdate
	mu             sync.RWMutex
}

type GameUpdate struct {
	Pet     Pet
	Message string
}

var gameState *GameState

func InitGame() {
	gameState = &GameState{
		Pet: Pet{
			Name:      "Jankypet",
			Health:    50,
			MaxHealth: 100,
			Mood:      "neutral",
			LastFed:   time.Now(),
			LastPet:   time.Now(),
			Actions:   make([]Action, 0),
		},
		ConnectedUsers: make(map[string]bool),
		Subscribers:    make(map[string]chan GameUpdate),
	}

	// Start health decay goroutine
	go healthDecayLoop()
}

func healthDecayLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		gameState.mu.Lock()
		if gameState.Pet.Health > 0 {
			gameState.Pet.Health--
			if gameState.Pet.Health <= 0 {
				gameState.Pet.Health = 0
				gameState.Pet.Mood = "dead"
				gameState.addAction("system", "died")
			} else if gameState.Pet.Health < 20 {
				gameState.Pet.Mood = "dying"
			} else if gameState.Pet.Health < 40 {
				gameState.Pet.Mood = "sad"
			}
		}
		pet := gameState.Pet
		gameState.mu.Unlock()

		broadcast(GameUpdate{Pet: pet, Message: ""})
	}
}

func Subscribe(userID string) chan GameUpdate {
	gameState.mu.Lock()
	defer gameState.mu.Unlock()

	ch := make(chan GameUpdate, 10)
	gameState.Subscribers[userID] = ch
	gameState.ConnectedUsers[userID] = true
	return ch
}

func Unsubscribe(userID string) {
	gameState.mu.Lock()
	defer gameState.mu.Unlock()

	if ch, exists := gameState.Subscribers[userID]; exists {
		close(ch)
		delete(gameState.Subscribers, userID)
	}
	delete(gameState.ConnectedUsers, userID)
}

func broadcast(update GameUpdate) {
	gameState.mu.RLock()
	defer gameState.mu.RUnlock()

	for _, ch := range gameState.Subscribers {
		select {
		case ch <- update:
		default:
			// Channel full, skip
		}
	}
}

func (gs *GameState) addAction(user, actionType string) {
	action := Action{
		User:      user,
		Type:      actionType,
		Timestamp: time.Now(),
	}
	gs.Pet.Actions = append(gs.Pet.Actions, action)

	// Keep only last 5 actions
	if len(gs.Pet.Actions) > 5 {
		gs.Pet.Actions = gs.Pet.Actions[len(gs.Pet.Actions)-5:]
	}
}

func FeedPet(user string) string {
	gameState.mu.Lock()

	if gameState.Pet.Health <= 0 {
		gameState.Pet.Health = 30
		gameState.Pet.Mood = "revived"
		gameState.addAction(user, "revived")
		pet := gameState.Pet
		gameState.mu.Unlock()
		broadcast(GameUpdate{Pet: pet, Message: fmt.Sprintf("%s revived Jankypet!", user)})
		return "You revived Jankypet!"
	}

	gameState.Pet.Health += 15
	if gameState.Pet.Health > gameState.Pet.MaxHealth {
		gameState.Pet.Health = gameState.Pet.MaxHealth
	}
	gameState.Pet.LastFed = time.Now()
	gameState.Pet.Mood = "happy"
	gameState.addAction(user, "fed")

	pet := gameState.Pet
	gameState.mu.Unlock()
	broadcast(GameUpdate{Pet: pet, Message: fmt.Sprintf("%s fed Jankypet!", user)})
	return "You fed Jankypet!"
}

func PetPet(user string) string {
	gameState.mu.Lock()

	if gameState.Pet.Health <= 0 {
		gameState.mu.Unlock()
		return "Can't pet a dead pet!"
	}

	gameState.Pet.Health += 5
	if gameState.Pet.Health > gameState.Pet.MaxHealth {
		gameState.Pet.Health = gameState.Pet.MaxHealth
	}
	gameState.Pet.LastPet = time.Now()
	if gameState.Pet.Mood != "happy" {
		gameState.Pet.Mood = "content"
	}
	gameState.addAction(user, "petted")

	pet := gameState.Pet
	gameState.mu.Unlock()
	broadcast(GameUpdate{Pet: pet, Message: fmt.Sprintf("%s petted Jankypet!", user)})
	return "You petted Jankypet!"
}

func HitPet(user string) string {
	gameState.mu.Lock()

	if gameState.Pet.Health <= 0 {
		gameState.mu.Unlock()
		return "Stop! It's already dead!"
	}

	gameState.Pet.Health -= 20
	if gameState.Pet.Health < 0 {
		gameState.Pet.Health = 0
	}

	var message string
	if gameState.Pet.Health <= 0 {
		gameState.Pet.Mood = "dead"
		gameState.addAction(user, "killed")
		message = "You killed Jankypet! 💀"
	} else {
		gameState.Pet.Mood = "angry"
		gameState.addAction(user, "hit")
		message = "You hit Jankypet! 😢"
	}

	pet := gameState.Pet
	gameState.mu.Unlock()
	broadcast(GameUpdate{Pet: pet, Message: fmt.Sprintf("%s hit Jankypet!", user)})
	return message
}

func GetGameState() Pet {
	gameState.mu.RLock()
	defer gameState.mu.RUnlock()
	return gameState.Pet
}

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// User struct to store registration information
type User struct {
	Username string
	Password string
}

// In-memory user store
var userStore = make(map[string]User)
var storeMutex sync.Mutex

// Flag to check if streaming is active (per user, if needed)
var isStreaming = make(map[string]bool)

// Registration handler
func registrationHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	if username == "" || password == "" {
		http.Error(w, "Missing username or password", http.StatusBadRequest)
		return
	}

	// Lock to avoid concurrent map writes
	storeMutex.Lock()
	userStore[username] = User{Username: username, Password: password}
	isStreaming[username] = false // Initialize streaming status
	storeMutex.Unlock()

	fmt.Fprintf(w, "User %s registered successfully\n", username)
}

// Start stream handler to receive audio from client-side JavaScript
func startStreamHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	// Authenticate user
	storeMutex.Lock()
	user, exists := userStore[username]
	storeMutex.Unlock()

	if !exists || user.Password != password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Mark user as streaming
	storeMutex.Lock()
	isStreaming[username] = true
	storeMutex.Unlock()

	// Open the audio file in append mode to continuously add audio data
	file, err := os.OpenFile(username+".wav", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Failed to open audio file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Write the incoming audio stream chunk to the file
	_, err = io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, "Failed to write audio data", http.StatusInternalServerError)
		return
	}
	//fmt.Println("Audio chunk received.")
	fmt.Fprintf(w, "Audio chunk received from %s\n", username)
}

// Stop stream handler to stop receiving audio from client-side
func stopStreamHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	// Authenticate user
	storeMutex.Lock()
	user, exists := userStore[username]
	storeMutex.Unlock()

	if !exists || user.Password != password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Mark user as not streaming
	storeMutex.Lock()
	isStreaming[username] = false
	storeMutex.Unlock()
	fmt.Fprintf(w, "Streaming stopped for %s\n", username)
}

// Listen stream handler
func listenStreamHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the recorded audio file as the stream
	username := r.URL.Query().Get("username")
	file, err := os.Open(username + ".wav")
	if err != nil {
		http.Error(w, "Audio stream not available", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Content-Disposition", "inline;filename=audio-stream.wav")
	//fmt.Printf("listen audio chunk call recvd")

	// Serve the audio file in chunks
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to stream audio", http.StatusInternalServerError)
		return
	}
}

func main() {

	// Serve on local network within WiFi range
	http.HandleFunc("/register", registrationHandler)
	http.HandleFunc("/startstream", startStreamHandler)
	http.HandleFunc("/stopstream", stopStreamHandler)
	http.HandleFunc("/listenstream", listenStreamHandler)

	// Serve the HTML page that captures audio
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("Server started. Listening on http://localhost:8080/")

	http.ListenAndServe(":8080", nil)
}

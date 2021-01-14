// Package state provides the application state. Currently this is only the
// open lobbies. However, the lobby state itself is managed in the game
// package. On top of this, we automatically clean up deserted lobbies
// in this package, as it is much easier in a centralized places and also
// protects us from flooding the server with goroutines.
package state

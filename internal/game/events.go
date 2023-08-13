package game

// Eevnts that are just incomming from the client.
const (
	EventTypeStart          = "start"
	EventTypeRequestDrawing = "request-drawing"
	EventTypeChooseWord     = "choose-word"
	EventTypeUndo           = "undo"
)

// Events that are outgoing only.
const (
	EventTypeUpdatePlayers            = "update-players"
	EventTypeUpdateWordHint           = "update-wordhint"
	EventTypeCorrectGuess             = "correct-guess"
	EventTypeCloseGuess               = "close-guess"
	EventTypeSystemMessage            = "system-message"
	EventTypeNonGuessingPlayerMessage = "non-guessing-player-message"
	EventTypeReady                    = "ready"
	EventTypeGameOver                 = "game-over"
	EventTypeYourTurn                 = "your-turn"
	EventTypeNextTurn                 = "next-turn"
	EventTypeDrawing                  = "drawing"
	EventTypeDrawerKicked             = "drawer-kicked"
	EventTypeOwnerChange              = "owner-change"
	EventTypeLobbySettingsChanged     = "lobby-settings-changed"
	EventTypeShutdown                 = "shutdown"
)

// Events that are bidirectional.
var (
	EventTypeKickVote          = "kick-vote"
	EventTypeNameChange        = "name-change"
	EventTypeMessage           = "message"
	EventTypeLine              = "line"
	EventTypeFill              = "fill"
	EventTypeClearDrawingBoard = "clear-drawing-board"
)

package errors

import "fmt"

type SpotError struct {
	Message string
	Err     string
}

func (e *SpotError) Error() string {
	if e.Err != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Err)
	}
	return e.Message
}

// NewSpotError creates a *SpotError containing the provided message and underlying error text.
// If err is empty, the returned SpotError will report only the message when formatted.
func NewSpotError(message string, err string) *SpotError {
	return &SpotError{Message: message, Err: err}
}

type CaptchaException struct{ *SpotError }
type SolverError struct{ *SpotError }
type LoginError struct{ *SpotError }
type UserError struct{ *SpotError }
type PlaylistError struct{ *SpotError }
type SaverError struct{ *SpotError }
type SongError struct{ *SpotError }
type ArtistError struct{ *SpotError }
type BaseClientError struct{ *SpotError }
type RequestError struct{ *SpotError }
type GeneratorError struct{ *SpotError }
type PasswordError struct{ *SpotError }
type FamilyError struct{ *SpotError }
type WebSocketError struct{ *SpotError }
type PlayerError struct{ *SpotError }
type AlbumError struct{ *SpotError }
type PodcastError struct{ *SpotError }

// NewLoginError creates a LoginError that wraps a SpotError with the provided message and underlying error string.
// The msg parameter is the primary error message; err is an optional underlying error string.
func NewLoginError(msg, err string) *LoginError { return &LoginError{NewSpotError(msg, err)} }
// NewRequestError creates a RequestError containing the given message and underlying error string.
func NewRequestError(msg, err string) *RequestError { return &RequestError{NewSpotError(msg, err)} }
// NewBaseClientError creates a BaseClientError carrying the provided message and underlying error string.
// The resulting error's Error() returns "message: err" when the underlying error string is non-empty, otherwise it returns the message.
func NewBaseClientError(msg, err string) *BaseClientError { return &BaseClientError{NewSpotError(msg, err)} }
// NewPlaylistError creates a PlaylistError that wraps a SpotError initialized with the given message and underlying error string.
// The returned *PlaylistError carries the provided message and optional underlying error text for formatting by Error().
func NewPlaylistError(msg, err string) *PlaylistError { return &PlaylistError{NewSpotError(msg, err)} }
// NewUserError creates a UserError that wraps a SpotError initialized with the provided message and underlying error string.
func NewUserError(msg, err string) *UserError { return &UserError{NewSpotError(msg, err)} }
// NewSongError creates a *SongError containing the provided message and underlying error string.
// The err parameter may be empty if there is no underlying error.
func NewSongError(msg, err string) *SongError { return &SongError{NewSpotError(msg, err)} }
// NewArtistError creates an ArtistError containing the given message and underlying error string.
// The error's string representation includes the underlying error after the message when the underlying error is non-empty.
func NewArtistError(msg, err string) *ArtistError { return &ArtistError{NewSpotError(msg, err)} }
// NewAlbumError creates an AlbumError that wraps a SpotError containing the given message and underlying error string.
func NewAlbumError(msg, err string) *AlbumError { return &AlbumError{NewSpotError(msg, err)} }
// NewPodcastError creates a PodcastError that wraps a SpotError initialized with the provided message and underlying error string.
func NewPodcastError(msg, err string) *PodcastError { return &PodcastError{NewSpotError(msg, err)} }
// NewWebSocketError creates a WebSocketError with the provided message and underlying error string.
func NewWebSocketError(msg, err string) *WebSocketError { return &WebSocketError{NewSpotError(msg, err)} }
// NewPlayerError creates a PlayerError containing a SpotError with the given message and underlying error string.
func NewPlayerError(msg, err string) *PlayerError { return &PlayerError{NewSpotError(msg, err)} }

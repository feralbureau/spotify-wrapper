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

func NewLoginError(msg, err string) *LoginError     { return &LoginError{NewSpotError(msg, err)} }
func NewRequestError(msg, err string) *RequestError { return &RequestError{NewSpotError(msg, err)} }
func NewBaseClientError(msg, err string) *BaseClientError {
	return &BaseClientError{NewSpotError(msg, err)}
}
func NewPlaylistError(msg, err string) *PlaylistError { return &PlaylistError{NewSpotError(msg, err)} }
func NewUserError(msg, err string) *UserError         { return &UserError{NewSpotError(msg, err)} }
func NewSongError(msg, err string) *SongError         { return &SongError{NewSpotError(msg, err)} }
func NewArtistError(msg, err string) *ArtistError     { return &ArtistError{NewSpotError(msg, err)} }
func NewAlbumError(msg, err string) *AlbumError       { return &AlbumError{NewSpotError(msg, err)} }
func NewPodcastError(msg, err string) *PodcastError   { return &PodcastError{NewSpotError(msg, err)} }
func NewWebSocketError(msg, err string) *WebSocketError {
	return &WebSocketError{NewSpotError(msg, err)}
}
func NewPlayerError(msg, err string) *PlayerError { return &PlayerError{NewSpotError(msg, err)} }

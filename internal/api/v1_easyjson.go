// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package api

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	game "github.com/scribble-rs/scribble.rs/internal/game"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi(in *jlexer.Lexer, out *LobbyEntry) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "lobbyId":
			out.LobbyID = string(in.String())
		case "wordpack":
			out.Wordpack = string(in.String())
		case "scoring":
			out.Scoring = string(in.String())
		case "state":
			out.State = game.State(in.String())
		case "playerCount":
			out.PlayerCount = int(in.Int())
		case "maxPlayers":
			out.MaxPlayers = int(in.Int())
		case "round":
			out.Round = int(in.Int())
		case "rounds":
			out.Rounds = int(in.Int())
		case "drawingTime":
			out.DrawingTime = int(in.Int())
		case "maxClientsPerIp":
			out.MaxClientsPerIP = int(in.Int())
		case "customWords":
			out.CustomWords = bool(in.Bool())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi(out *jwriter.Writer, in LobbyEntry) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"lobbyId\":"
		out.RawString(prefix[1:])
		out.String(string(in.LobbyID))
	}
	{
		const prefix string = ",\"wordpack\":"
		out.RawString(prefix)
		out.String(string(in.Wordpack))
	}
	{
		const prefix string = ",\"scoring\":"
		out.RawString(prefix)
		out.String(string(in.Scoring))
	}
	{
		const prefix string = ",\"state\":"
		out.RawString(prefix)
		out.String(string(in.State))
	}
	{
		const prefix string = ",\"playerCount\":"
		out.RawString(prefix)
		out.Int(int(in.PlayerCount))
	}
	{
		const prefix string = ",\"maxPlayers\":"
		out.RawString(prefix)
		out.Int(int(in.MaxPlayers))
	}
	{
		const prefix string = ",\"round\":"
		out.RawString(prefix)
		out.Int(int(in.Round))
	}
	{
		const prefix string = ",\"rounds\":"
		out.RawString(prefix)
		out.Int(int(in.Rounds))
	}
	{
		const prefix string = ",\"drawingTime\":"
		out.RawString(prefix)
		out.Int(int(in.DrawingTime))
	}
	{
		const prefix string = ",\"maxClientsPerIp\":"
		out.RawString(prefix)
		out.Int(int(in.MaxClientsPerIP))
	}
	{
		const prefix string = ",\"customWords\":"
		out.RawString(prefix)
		out.Bool(bool(in.CustomWords))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v LobbyEntry) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v LobbyEntry) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *LobbyEntry) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *LobbyEntry) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi(l, v)
}
func easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi1(in *jlexer.Lexer, out *LobbyEntries) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		in.Skip()
		*out = nil
	} else {
		in.Delim('[')
		if *out == nil {
			if !in.IsDelim(']') {
				*out = make(LobbyEntries, 0, 8)
			} else {
				*out = LobbyEntries{}
			}
		} else {
			*out = (*out)[:0]
		}
		for !in.IsDelim(']') {
			var v1 *LobbyEntry
			if in.IsNull() {
				in.Skip()
				v1 = nil
			} else {
				if v1 == nil {
					v1 = new(LobbyEntry)
				}
				(*v1).UnmarshalEasyJSON(in)
			}
			*out = append(*out, v1)
			in.WantComma()
		}
		in.Delim(']')
	}
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi1(out *jwriter.Writer, in LobbyEntries) {
	if in == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v2, v3 := range in {
			if v2 > 0 {
				out.RawByte(',')
			}
			if v3 == nil {
				out.RawString("null")
			} else {
				(*v3).MarshalEasyJSON(out)
			}
		}
		out.RawByte(']')
	}
}

// MarshalJSON supports json.Marshaler interface
func (v LobbyEntries) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v LobbyEntries) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *LobbyEntries) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *LobbyEntries) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi1(l, v)
}
func easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi2(in *jlexer.Lexer, out *LobbyData) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "lobbyId":
			out.LobbyID = string(in.String())
		case "drawingBoardBaseWidth":
			out.DrawingBoardBaseWidth = int(in.Int())
		case "drawingBoardBaseHeight":
			out.DrawingBoardBaseHeight = int(in.Int())
		case "minBrushSize":
			out.MinBrushSize = int(in.Int())
		case "maxBrushSize":
			out.MaxBrushSize = int(in.Int())
		case "canvasColor":
			out.CanvasColor = uint8(in.Uint8())
		case "suggestedBrushSizes":
			if in.IsNull() {
				in.Skip()
			} else {
				copy(out.SuggestedBrushSizes[:], in.Bytes())
			}
		case "public":
			out.Public = bool(in.Bool())
		case "maxPlayers":
			out.MaxPlayers = int(in.Int())
		case "customWordsPerTurn":
			out.CustomWordsPerTurn = int(in.Int())
		case "clientsPerIpLimit":
			out.ClientsPerIPLimit = int(in.Int())
		case "rounds":
			out.Rounds = int(in.Int())
		case "drawingTime":
			out.DrawingTime = int(in.Int())
		case "minDrawingTime":
			out.MinDrawingTime = int(in.Int())
		case "maxDrawingTime":
			out.MaxDrawingTime = int(in.Int())
		case "minRounds":
			out.MinRounds = int(in.Int())
		case "maxRounds":
			out.MaxRounds = int(in.Int())
		case "minMaxPlayers":
			out.MinMaxPlayers = int(in.Int())
		case "maxMaxPlayers":
			out.MaxMaxPlayers = int(in.Int())
		case "minClientsPerIpLimit":
			out.MinClientsPerIPLimit = int(in.Int())
		case "maxClientsPerIpLimit":
			out.MaxClientsPerIPLimit = int(in.Int())
		case "minCustomWordsPerTurn":
			out.MinCustomWordsPerTurn = int(in.Int())
		case "maxCustomWordsPerTurn":
			out.MaxCustomWordsPerTurn = int(in.Int())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi2(out *jwriter.Writer, in LobbyData) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"lobbyId\":"
		out.RawString(prefix[1:])
		out.String(string(in.LobbyID))
	}
	{
		const prefix string = ",\"drawingBoardBaseWidth\":"
		out.RawString(prefix)
		out.Int(int(in.DrawingBoardBaseWidth))
	}
	{
		const prefix string = ",\"drawingBoardBaseHeight\":"
		out.RawString(prefix)
		out.Int(int(in.DrawingBoardBaseHeight))
	}
	{
		const prefix string = ",\"minBrushSize\":"
		out.RawString(prefix)
		out.Int(int(in.MinBrushSize))
	}
	{
		const prefix string = ",\"maxBrushSize\":"
		out.RawString(prefix)
		out.Int(int(in.MaxBrushSize))
	}
	{
		const prefix string = ",\"canvasColor\":"
		out.RawString(prefix)
		out.Uint8(uint8(in.CanvasColor))
	}
	{
		const prefix string = ",\"suggestedBrushSizes\":"
		out.RawString(prefix)
		out.Base64Bytes(in.SuggestedBrushSizes[:])
	}
	{
		const prefix string = ",\"public\":"
		out.RawString(prefix)
		out.Bool(bool(in.Public))
	}
	{
		const prefix string = ",\"maxPlayers\":"
		out.RawString(prefix)
		out.Int(int(in.MaxPlayers))
	}
	{
		const prefix string = ",\"customWordsPerTurn\":"
		out.RawString(prefix)
		out.Int(int(in.CustomWordsPerTurn))
	}
	{
		const prefix string = ",\"clientsPerIpLimit\":"
		out.RawString(prefix)
		out.Int(int(in.ClientsPerIPLimit))
	}
	{
		const prefix string = ",\"rounds\":"
		out.RawString(prefix)
		out.Int(int(in.Rounds))
	}
	{
		const prefix string = ",\"drawingTime\":"
		out.RawString(prefix)
		out.Int(int(in.DrawingTime))
	}
	{
		const prefix string = ",\"minDrawingTime\":"
		out.RawString(prefix)
		out.Int(int(in.MinDrawingTime))
	}
	{
		const prefix string = ",\"maxDrawingTime\":"
		out.RawString(prefix)
		out.Int(int(in.MaxDrawingTime))
	}
	{
		const prefix string = ",\"minRounds\":"
		out.RawString(prefix)
		out.Int(int(in.MinRounds))
	}
	{
		const prefix string = ",\"maxRounds\":"
		out.RawString(prefix)
		out.Int(int(in.MaxRounds))
	}
	{
		const prefix string = ",\"minMaxPlayers\":"
		out.RawString(prefix)
		out.Int(int(in.MinMaxPlayers))
	}
	{
		const prefix string = ",\"maxMaxPlayers\":"
		out.RawString(prefix)
		out.Int(int(in.MaxMaxPlayers))
	}
	{
		const prefix string = ",\"minClientsPerIpLimit\":"
		out.RawString(prefix)
		out.Int(int(in.MinClientsPerIPLimit))
	}
	{
		const prefix string = ",\"maxClientsPerIpLimit\":"
		out.RawString(prefix)
		out.Int(int(in.MaxClientsPerIPLimit))
	}
	{
		const prefix string = ",\"minCustomWordsPerTurn\":"
		out.RawString(prefix)
		out.Int(int(in.MinCustomWordsPerTurn))
	}
	{
		const prefix string = ",\"maxCustomWordsPerTurn\":"
		out.RawString(prefix)
		out.Int(int(in.MaxCustomWordsPerTurn))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v LobbyData) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v LobbyData) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *LobbyData) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *LobbyData) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi2(l, v)
}
func easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi3(in *jlexer.Lexer, out *Gallery) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		in.Skip()
		*out = nil
	} else {
		in.Delim('[')
		if *out == nil {
			if !in.IsDelim(']') {
				*out = make(Gallery, 0, 1)
			} else {
				*out = Gallery{}
			}
		} else {
			*out = (*out)[:0]
		}
		for !in.IsDelim(']') {
			var v6 game.GalleryDrawing
			easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalGame(in, &v6)
			*out = append(*out, v6)
			in.WantComma()
		}
		in.Delim(']')
	}
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi3(out *jwriter.Writer, in Gallery) {
	if in == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v7, v8 := range in {
			if v7 > 0 {
				out.RawByte(',')
			}
			easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalGame(out, v8)
		}
		out.RawByte(']')
	}
}

// MarshalJSON supports json.Marshaler interface
func (v Gallery) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Gallery) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalApi3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Gallery) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Gallery) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalApi3(l, v)
}
func easyjson102f8a2fDecodeGithubComScribbleRsScribbleRsInternalGame(in *jlexer.Lexer, out *game.GalleryDrawing) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "word":
			out.Word = string(in.String())
		case "events":
			if in.IsNull() {
				in.Skip()
				out.Events = nil
			} else {
				in.Delim('[')
				if out.Events == nil {
					if !in.IsDelim(']') {
						out.Events = make([]interface{}, 0, 4)
					} else {
						out.Events = []interface{}{}
					}
				} else {
					out.Events = (out.Events)[:0]
				}
				for !in.IsDelim(']') {
					var v9 interface{}
					if m, ok := v9.(easyjson.Unmarshaler); ok {
						m.UnmarshalEasyJSON(in)
					} else if m, ok := v9.(json.Unmarshaler); ok {
						_ = m.UnmarshalJSON(in.Raw())
					} else {
						v9 = in.Interface()
					}
					out.Events = append(out.Events, v9)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson102f8a2fEncodeGithubComScribbleRsScribbleRsInternalGame(out *jwriter.Writer, in game.GalleryDrawing) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"word\":"
		out.RawString(prefix[1:])
		out.String(string(in.Word))
	}
	{
		const prefix string = ",\"events\":"
		out.RawString(prefix)
		if in.Events == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v10, v11 := range in.Events {
				if v10 > 0 {
					out.RawByte(',')
				}
				if m, ok := v11.(easyjson.Marshaler); ok {
					m.MarshalEasyJSON(out)
				} else if m, ok := v11.(json.Marshaler); ok {
					out.Raw(m.MarshalJSON())
				} else {
					out.Raw(json.Marshal(v11))
				}
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

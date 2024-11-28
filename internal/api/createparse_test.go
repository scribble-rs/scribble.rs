package api

import (
	"reflect"
	"testing"

	"github.com/scribble-rs/scribble.rs/internal/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Test_parsePlayerName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{"empty name", "", "", true},
		{"blank name", " ", "", true},
		{"one letter name", "a", "a", false},
		{"normal name", "Scribble", "Scribble", false},
		{"name with space in the middle", "Hello World", "Hello World", false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParsePlayerName(testCase.value)
			if (err != nil) != testCase.wantErr {
				t.Errorf("parsePlayerName() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if got != testCase.want {
				t.Errorf("parsePlayerName() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func Test_parseDrawingTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{"empty value", "", 0, true},
		{"space", " ", 0, true},
		{"less than minimum", "59", 0, true},
		{"more than maximum", "301", 0, true},
		{"maximum", "300", 300, false},
		{"minimum", "60", 60, false},
		{"something valid", "150", 150, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseDrawingTime(&config.Default, testCase.value)
			if (err != nil) != testCase.wantErr {
				t.Errorf("parseDrawingTime() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if got != testCase.want {
				t.Errorf("parseDrawingTime() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func Test_parseRounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{"empty value", "", 0, true},
		{"space", " ", 0, true},
		{"less than minimum", "0", 0, true},
		{"more than maximum", "21", 0, true},
		{"maximum", "20", 20, false},
		{"minimum", "1", 1, false},
		{"something valid", "15", 15, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseRounds(&config.Default, testCase.value)
			if (err != nil) != testCase.wantErr {
				t.Errorf("parseRounds() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if got != testCase.want {
				t.Errorf("parseRounds() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func Test_parseMaxPlayers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{"empty value", "", 0, true},
		{"space", " ", 0, true},
		{"less than minimum", "1", 0, true},
		{"more than maximum", "25", 0, true},
		{"maximum", "24", 24, false},
		{"minimum", "2", 2, false},
		{"something valid", "15", 15, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseMaxPlayers(&config.Default, testCase.value)
			if (err != nil) != testCase.wantErr {
				t.Errorf("parseMaxPlayers() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if got != testCase.want {
				t.Errorf("parseMaxPlayers() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func Test_parseCustomWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    []string
		wantErr bool
	}{
		{"emtpty", "", nil, false},
		{"spaces", "   ", nil, false},
		{"spaces with comma in middle", "  , ", nil, true},
		{"single word", "hello", []string{"hello"}, false},
		{"single word upper to lower", "HELLO", []string{"hello"}, false},
		{"single word with spaces around", "   hello ", []string{"hello"}, false},
		{"two words", "hello,world", []string{"hello", "world"}, false},
		{"two words with spaces around", " hello , world ", []string{"hello", "world"}, false},
		{"sentence and word", "What a great day, hello ", []string{"what a great day", "hello"}, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseCustomWords(cases.Lower(language.English), testCase.value)
			if (err != nil) != testCase.wantErr {
				t.Errorf("parseCustomWords() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("parseCustomWords() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func Test_parseCustomWordsPerTurn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{"empty value", "", 0, true},
		{"space", " ", 0, true},
		{"lesss than minimum, zero", "0", 0, true},
		{"less than minimum, negative", "-1", 0, true},
		{"more than maximum", "4", 0, true},
		{"minimum", "1", 1, false},
		{"maximum", "3", 3, false},
		{"something valid", "2", 2, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseCustomWordsPerTurn(testCase.value)
			if (err != nil) != testCase.wantErr {
				t.Errorf("parseCustomWordsPerTurn() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if got != testCase.want {
				t.Errorf("parseCustomWordsPerTurn() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func Test_parseBoolean(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    bool
		wantErr bool
	}{
		{"empty value", "", false, false},
		{"space", " ", false, true},
		{"garbage", "garbage", false, true},
		{"true", "true", true, false},
		{"true upper", "TRUE", true, false},
		{"true mixed casing", "TruE", true, false},
		{"false", "false", false, false},
		{"false upper", "FALSE", false, false},
		{"false mixed casing", "FalsE", false, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseBoolean("name", testCase.value)
			if (err != nil) != testCase.wantErr {
				t.Errorf("parseBoolean() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if got != testCase.want {
				t.Errorf("parseBoolean() = %v, want %v", got, testCase.want)
			}
		})
	}
}

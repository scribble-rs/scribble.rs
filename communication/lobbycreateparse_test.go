package communication

import (
	"reflect"
	"testing"
)

func Test_parsePlayerName(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePlayerName(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePlayerName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parsePlayerName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePassword(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{"empty password", "", "", false},
		{"space as password", " ", " ", false},
		{"word as password", "word", "word", false},
		{"string with space in the middle", "Hello World", "Hello World", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePassword(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parsePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseDrawingTime(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDrawingTime(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDrawingTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseDrawingTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseRounds(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRounds(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRounds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseRounds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseMaxPlayers(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMaxPlayers(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMaxPlayers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseMaxPlayers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseCustomWords(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCustomWords(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCustomWords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseCustomWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseCustomWordChance(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{"empty value", "", 0, true},
		{"space", " ", 0, true},
		{"less than minimum", "-1", 0, true},
		{"more than maximum", "101", 0, true},
		{"maximum", "100", 100, false},
		{"minimum", "0", 0, false},
		{"something valid", "60", 60, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCustomWordsChance(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCustomWordsChance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseCustomWordsChance() = %v, want %v", got, tt.want)
			}
		})
	}
}

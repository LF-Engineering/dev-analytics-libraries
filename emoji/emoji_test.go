package emoji

import "testing"

func TestGetEmojiUnicode(t *testing.T) {
	type args struct {
		emojiStr string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Test 1",
			args{emojiStr: ":+1:"},
			"\U0001f44d",
		},
		{
			"Test 2",
			args{emojiStr: ":abacus:"},
			"\U0001f9ee",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEmojiUnicode(tt.args.emojiStr); got != tt.want {
				t.Errorf("getEmojiUnicode() = %v, want %v", got, tt.want)
			}
		})
	}
}

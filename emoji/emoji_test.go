package emoji

import "testing"

func Test_getEmojiUnicode(t *testing.T) {
	type args struct {
		emojiStr string
	}
	tests := []struct {
		name string
		args args
		want string
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEmojiUnicode(tt.args.emojiStr); got != tt.want {
				t.Errorf("getEmojiUnicode() = %v, want %v", got, tt.want)
			}
		})
	}
}
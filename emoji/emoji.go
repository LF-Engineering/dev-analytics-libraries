package emoji

func GetEmojiUnicode(emojiStr string) string {
	if _,ok := CodeMap[emojiStr]; ok {
		return CodeMap[emojiStr]
	}
	return emojiStr
}
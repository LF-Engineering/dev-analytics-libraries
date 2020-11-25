package emoji

func getEmojiUnicode(emojiStr string) string {
	if _,ok := CodeMap[emojiStr]; ok {
		return CodeMap[emojiStr]
	}
	return emojiStr
}
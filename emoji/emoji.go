package emoji

func getEmojiUnicode(emojiStr string) string {
	if _,ok := emojiCodeMap[emojiStr]; ok {
		return emojiCodeMap[emojiStr]
	}
	return emojiStr
}
package builder

func BuildTargetImage(repo, tag, digest string) string {
	if digest != "" {
		return repo + "@" + digest
	}
	return repo + ":" + tag
}

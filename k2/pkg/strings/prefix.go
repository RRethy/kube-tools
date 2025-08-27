package strings

func FindCommonPrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	for ci := 0; ci < len(paths[0]); ci++ {
		for _, p := range paths {
			if ci == len(p) || paths[0][ci] != p[ci] {
				return paths[0][:ci]
			}
		}
	}
	return paths[0]
}

package models

type SpyReader struct {
}

func (s *SpyReader) Read(p []byte) (n int, err error) {

	return 0, nil
}

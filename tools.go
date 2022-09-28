package toolkit

import "crypto/rand"

// characters for generating random string
const randStringSource = "abcdefhijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

// type used for instantiation.
// variables of this type will have access to methods with a *Tools reciever
type Tools struct {
}

// returns a string of random characters of length n, sourced from randStringSource
func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randStringSource)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

package fate

import (
	"math/rand"
	"strings"
	"sync/atomic"
	"testing"
)

// Make random corpuses of ngram data for scale testing.

func corpus(words []string, count int, sentlen func() int) []string {
	var sentences = make([]string, 0, count)
	for i := 0; i < count; i++ {
		sentences = append(sentences, sentence(words, sentlen()))
	}
	return sentences
}

func vocab(n int) []string {
	var words = make([]string, 0, n)
	for i := 0; i < n; i++ {
		words = append(words, randword(clamp(gauss(8, 5))))
	}
	return words
}

func gauss(mean, stddev float64) float64 {
	return rand.NormFloat64()*stddev + mean
}

func clamp(v float64) int {
	n := int(v + 0.5)
	if n < 1 {
		return 1
	}
	return n
}

var runes = []rune("abcdefghijklmnopqrstuvwxyz")

func randword(length int) string {
	var word = make([]rune, 0, length)
	for i := 0; i < length; i++ {
		word = append(word, runechoice(runes))
	}
	return string(word)
}

func runechoice(runes []rune) rune {
	return runes[rand.Intn(len(runes))]
}

func strchoice(strs []string) string {
	return strs[rand.Intn(len(strs))]
}

func sentence(words []string, n int) string {
	var ret = make([]string, 0, n)
	for i := 0; i < n; i++ {
		ret = append(ret, strchoice(words))
	}
	return strings.Join(ret, " ")
}

func BenchmarkLearn(b *testing.B) {
	sentences := corpus(vocab(100000), b.N, func() int {
		return clamp(gauss(10, 5))
	})

	model := NewModel(Config{})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sen := sentences[i%len(sentences)]
		model.Learn(sen)
		b.SetBytes(int64(len(sen)))
	}
}

func BenchmarkReply(b *testing.B) {
	sentences := corpus(vocab(100000), b.N, func() int {
		return clamp(gauss(10, 5))
	})

	model := NewModel(Config{})
	for _, sen := range sentences {
		model.Learn(sen)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		model.Reply(sentences[i%len(sentences)])
	}
}

func BenchmarkLearnParallel(b *testing.B) {
	sentences := corpus(vocab(100000), b.N, func() int {
		return clamp(gauss(10, 5))
	})

	model := NewModel(Config{})

	b.ReportAllocs()
	b.ResetTimer()

	var i int32
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			j := atomic.AddInt32(&i, 1)
			model.Learn(sentences[int(j)%len(sentences)])
		}
	})
}

func BenchmarkReplyParallel(b *testing.B) {
	sentences := corpus(vocab(100000), b.N, func() int {
		return clamp(gauss(10, 5))
	})

	model := NewModel(Config{})
	for _, sen := range sentences {
		model.Learn(sen)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var i int
		for pb.Next() {
			model.Reply(sentences[i%len(sentences)])
			i++
		}
	})
}

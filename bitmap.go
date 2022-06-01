package fate

import "github.com/RoaringBitmap/roaring"

type Bitmap struct {
	roar *roaring.Bitmap
}

func NewBitmap() *Bitmap {
	return &Bitmap{
		roar: roaring.New(),
	}
}

func (b *Bitmap) Len() int {
	return int(b.roar.GetCardinality())
}

func (b *Bitmap) Add(tok token) bool {
	if b.roar.Contains(uint32(tok)) {
		return false
	}
	b.roar.Add(uint32(tok))
	return true
}

func (b *Bitmap) Index(n int) token {
	v, err := b.roar.Select(uint32(n))
	if err != nil {
		panic(err)
	}
	return token(v)
}

func (b *Bitmap) Choice(r intn) token {
	return b.Index(r.Intn(b.Len()))
}

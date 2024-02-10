package delaywheel

type tWheel struct {
	tickMs    int64
	wheelSize int
	interval  int64
	buckets   []Bucket

	next *tWheel
}

func newWheel(tickMs int64, wheelSize int) *tWheel {

	buckets := make([]Bucket, wheelSize)
	for i := range buckets {
		buckets[i] = NewBucket()
	}

	return &tWheel{
		tickMs:    tickMs,
		wheelSize: wheelSize,
		interval:  tickMs * int64(wheelSize),
		buckets:   buckets,
	}
}

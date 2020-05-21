package stream

func ReportProgress(cnt Counter, prg chan<- int64) {
	for p := range cnt.Count() {
		prg <- p
	}
}

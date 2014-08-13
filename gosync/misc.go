package gosync

func newDoneChan(concurrent int) chan error {
	// Panic on any errors
	doneChan := make(chan error, concurrent)
	go func() {
		for {
			select {
			case err := <-doneChan:
				if err != nil {
					panic(err.Error())
				}
			}
		}
	}()
	return doneChan
}

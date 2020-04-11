package clock

import "time"

type (
	//Clock interface which allows time.Now() to be substituted with a fake time
	Clock interface {
		Now() time.Time
	}

	//RealClock implementation of the Clock interface with the real time.Now() attached
	RealClock struct{}
)

func (RealClock) Now() time.Time {
	return time.Now()
}

//NewClock returns the Clock interface with the actual time used via time.Now()
func NewClock() RealClock {
	return RealClock{}
}

//MockClock implementation of the Clock interface with a custom time
type MockClock struct {
	date time.Time
}

func (c MockClock) Now() time.Time {
	return c.date
}

//NewMockClock returns the Clock interface with a custom time
func NewMockClock(date time.Time) MockClock {
	return MockClock{
		date: date,
	}
}

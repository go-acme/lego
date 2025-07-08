package exec

import "github.com/stretchr/testify/mock"

type LogRecorder struct {
	mock.Mock
}

func (*LogRecorder) Fatal(args ...any) {
	panic("implement me")
}

func (*LogRecorder) Fatalln(args ...any) {
	panic("implement me")
}

func (*LogRecorder) Fatalf(format string, args ...any) {
	panic("implement me")
}

func (*LogRecorder) Print(args ...any) {
	panic("implement me")
}

func (l *LogRecorder) Println(args ...any) {
	l.Called(args...)
}

func (*LogRecorder) Printf(format string, args ...any) {
	panic("implement me")
}
